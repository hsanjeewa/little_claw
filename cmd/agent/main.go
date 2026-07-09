package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/config"
	"github.com/devops/agent/internal/infrastructure/database"
	"github.com/devops/agent/internal/infrastructure/inventory"
	"github.com/devops/agent/internal/infrastructure/llm"
	"github.com/devops/agent/internal/infrastructure/security"
	"github.com/devops/agent/internal/infrastructure/simulator"
	"github.com/devops/agent/internal/infrastructure/ssh"
	"github.com/devops/agent/internal/ui/tui"
)

// defaultMasterSecret is the dev fallback vault master key. In production,
// override it by setting LITTLE_CLAW_MASTER_KEY to a secret from your secret
// manager. The key is SHA-256 derived so any length is accepted.
const defaultMasterSecret = "a-very-secret-key-32-bytes-long!"

// masterKeyEnv is the environment variable holding the vault master key.
const masterKeyEnv = "LITTLE_CLAW_MASTER_KEY"

func main() {
	// Headless one-shot command to populate the encrypted vault (vault.enc)
	// from the environment (OPENAI_API_KEY, SUDO_PASS_*) without launching the
	// TUI. Usage: LITTLE_CLAW_MASTER_KEY=... ./small_claw init-vault
	if len(os.Args) > 1 && os.Args[1] == "init-vault" {
		loadEnv()
		if err := initVault(resolveMasterKey()); err != nil {
			log.Fatalf("vault init failed: %v", err)
		}
		fmt.Printf("vault initialized at %s\n", security.VaultPath())
		return
	}

	loadEnv()

	cfg, err := config.LoadConfig("config/config.toml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	repo, err := database.NewSQLiteRepository(cfg.Agent.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	sshTimeout := time.Duration(cfg.Agent.SSHTimeoutSeconds) * time.Second
	llmTimeout := time.Duration(cfg.LLM.TimeoutSeconds) * time.Second
	watchtowerRefreshInterval := time.Duration(cfg.Agent.WatchtowerRefreshIntervalSeconds) * time.Second

	masterKey := resolveMasterKey()
	vault := security.NewLocalEncryptedVault(masterKey)

	// Load any previously persisted (encrypted) secrets so they survive
	// restarts, then seed from the environment so a first run with
	// OPENAI_API_KEY / SUDO_PASS_* in .env persists them encrypted to vault.enc.
	if err := vault.Load(security.VaultPath()); err != nil {
		log.Fatalf("Failed to load vault: %v", err)
	}

	// Seed the LLM API key and per-host sudo passwords from the environment
	// into the encrypted vault (Ansible-style: encrypted at rest, resolved from
	// the vault rather than read directly from .env).
	security.ApplyAPIKeyEnv(vault, os.Environ())
	security.ApplySudoPasswordEnv(vault, os.Environ())

	// Resolve the LLM API key from the encrypted vault (never from plaintext
	// .env directly). It must have been seeded from OPENAI_API_KEY on a prior
	// run and persisted to vault.enc.
	apiKey, err := vault.GetAPIKey()
	if err != nil {
		log.Fatalf("OPENAI_API_KEY is not configured. Set OPENAI_API_KEY in .env (or the environment) for one run so it is stored encrypted in %s, then remove it from .env.", security.VaultPath())
	}

	analyzer := llm.NewLocalOpenAIClient(cfg.LLM.BaseURL, apiKey, cfg.LLM.Model)

	autopilotLLMClient := tui.NewLLMClient(tui.LLMConfig{
		BaseURL: cfg.LLM.BaseURL,
		APIKey:  apiKey,
		Model:   cfg.LLM.Model,
	})

	// Make plan generation context-aware about privilege escalation: a host with
	// a configured sudo password escalates via sudo (password auto-supplied);
	// otherwise it relies on passwordless sudo. The LLM is told per host whether
	// to include "sudo" or not.
	autopilotLLMClient = autopilotLLMClient.WithSudoPolicy(func(alias string) tui.SudoMode {
		if _, err := vault.GetSudoPassword(alias); err == nil {
			return tui.SudoPassword
		}
		return tui.SudoPasswordless
	})

	privateKeyData, err := os.ReadFile("test_keys/id_ed25519")
	if err == nil {
		_ = vault.EncryptAndStore("web-prod-01", string(privateKeyData))
		_ = vault.EncryptAndStore("db-master", string(privateKeyData))
	} else {
		log.Printf("Warning: test_keys/id_ed25519 not found. SSH execution will fail.")
	}

	// Persist the encrypted vault (API key + SSH/sudo secrets) to disk.
	if err := vault.Save(security.VaultPath()); err != nil {
		log.Fatalf("Failed to persist vault: %v", err)
	}

	sshClient := ssh.NewSSHClient(vault)
	idempHelper := ssh.NewLinuxIdempotencyAnalyzer(sshClient)

	engine := agent.NewEngine(sshClient, repo, analyzer, idempHelper, taskChan, logChan, sshTimeout, llmTimeout)

	watchtowerBackend := cfg.Agent.WatchtowerBackend
	var (
		targets          []inventory.TargetHost
		memoryCollector  tui.MemorySnapshotCollector
		cpuCollector     tui.CPUSnapshotCollector
		storageCollector tui.StorageSnapshotCollector
		networkCollector tui.NetworkSnapshotCollector
	)

	if watchtowerBackend == "simulator" {
		simBackend := simulator.NewWatchtowerBackend()
		targets = simBackend.Fleet()
		memoryCollector = simBackend.CollectMemory
		cpuCollector = simBackend.CollectCPU
		storageCollector = simBackend.CollectStorage
		networkCollector = simBackend.CollectNetwork
	} else {
		targets, err = inventory.LoadInventory(cfg.Agent.InventoryPath)
		if err != nil {
			log.Fatalf("Failed to load inventory: %v", err)
		}
		watchtowerCollector := ssh.NewWatchtowerMemoryCollector(sshClient)
		watchtowerCPUCollector := ssh.NewWatchtowerCPUCollector(sshClient)
		watchtowerStorageCollector := ssh.NewWatchtowerStorageCollector(sshClient)
		watchtowerNetworkCollector := ssh.NewWatchtowerNetworkCollector(sshClient)
		memoryCollector = watchtowerCollector.CollectMemory
		cpuCollector = watchtowerCPUCollector.CollectCPU
		storageCollector = watchtowerStorageCollector.CollectStorage
		networkCollector = watchtowerNetworkCollector.CollectNetwork
	}

	var tasks []agent.Task
	if watchtowerBackend != "simulator" {
		for _, target := range targets {
			if target.Alias == "db-master" {
				t, _ := agent.NewTask(uuid.New().String(), target.Alias, target.IP, target.Port, target.User, "pg_dump data.sql", true)
				tasks = append(tasks, t)
			} else {
				t1, _ := agent.NewTask(uuid.New().String(), target.Alias, target.IP, target.Port, target.User, "uptime", false)
				t2, _ := agent.NewTask(uuid.New().String(), target.Alias, target.IP, target.Port, target.User, "sudo apt-get update", true)
				tasks = append(tasks, t1, t2)
			}
		}
	}

	model := tui.NewShellWithInventoryAndAllCollectors(taskChan, logChan, hitlChan, tasks, targets, autopilotLLMClient, memoryCollector, cpuCollector, storageCollector, networkCollector, watchtowerRefreshInterval).WithExecutor(sshClient)

	if watchtowerBackend != "simulator" {
		go func() {
			for i := range tasks {
				t := tasks[i]
				time.Sleep(1 * time.Second) // Stagger execution slightly
				engine.RunTask(context.Background(), t, hitlChan)
			}
		}()
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// loadEnv loads .env, searching upward from the working directory so the file
// is found regardless of where the binary is launched from. godotenv.Overload
// ensures values in .env take precedence over any OPENAI_API_KEY (or other
// secret) already present in the process environment, which would otherwise
// shadow the correct key and cause authentication failures against OpenRouter.
func loadEnv() {
	_ = godotenv.Overload()

	dir, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Overload(path); err == nil {
				return
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

// resolveMasterKey returns the vault master key, preferring LITTLE_CLAW_MASTER_KEY
// from the environment and falling back to the dev default. The value is
// SHA-256 derived so any length is accepted and the key is always 32 bytes.
func resolveMasterKey() []byte {
	secret := os.Getenv(masterKeyEnv)
	if secret == "" {
		secret = defaultMasterSecret
	}
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// initVault seeds the encrypted vault from the environment and persists it to
// vault.enc. OPENAI_API_KEY and SUDO_PASS_* are read from the process
// environment (loaded from .env by loadEnv), and dev SSH keys are added when
// present. It does not launch the TUI.
func initVault(masterKey []byte) error {
	vault := security.NewLocalEncryptedVault(masterKey)

	// Merge with any existing vault so prior secrets are preserved.
	if err := vault.Load(security.VaultPath()); err != nil {
		return fmt.Errorf("load existing vault: %w", err)
	}

	security.ApplyAPIKeyEnv(vault, os.Environ())
	security.ApplySudoPasswordEnv(vault, os.Environ())

	if data, err := os.ReadFile("test_keys/id_ed25519"); err == nil {
		_ = vault.EncryptAndStore("web-prod-01", string(data))
		_ = vault.EncryptAndStore("db-master", string(data))
	}

	if err := vault.Save(security.VaultPath()); err != nil {
		return fmt.Errorf("save vault: %w", err)
	}
	return nil
}
