package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

func main() {
	_ = godotenv.Overload()

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

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatalf("OPENAI_API_KEY is not set. Please configure it in .env or environment.")
	}

	sshTimeout := time.Duration(cfg.Agent.SSHTimeoutSeconds) * time.Second
	llmTimeout := time.Duration(cfg.LLM.TimeoutSeconds) * time.Second
	watchtowerRefreshInterval := time.Duration(cfg.Agent.WatchtowerRefreshIntervalSeconds) * time.Second

	analyzer := llm.NewLocalOpenAIClient(cfg.LLM.BaseURL, apiKey, cfg.LLM.Model)

	autopilotLLMClient := tui.NewLLMClient(tui.LLMConfig{
		BaseURL: cfg.LLM.BaseURL,
		APIKey:  apiKey,
		Model:   cfg.LLM.Model,
	})

	masterSecret := []byte("a-very-secret-key-32-bytes-long!")
	vault := security.NewLocalEncryptedVault(masterSecret)

	privateKeyData, err := os.ReadFile("test_keys/id_ed25519")
	if err == nil {
		_ = vault.EncryptAndStore("web-prod-01", string(privateKeyData))
		_ = vault.EncryptAndStore("db-master", string(privateKeyData))
	} else {
		log.Printf("Warning: test_keys/id_ed25519 not found. SSH execution will fail.")
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
