package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/database"
	"github.com/devops/agent/internal/infrastructure/inventory"
	"github.com/devops/agent/internal/infrastructure/llm"
	"github.com/devops/agent/internal/infrastructure/security"
	"github.com/devops/agent/internal/infrastructure/ssh"
	"github.com/devops/agent/internal/ui/tui"
)

func main() {
	_ = godotenv.Load()

	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	dbPath := "./agent.db"
	repo, err := database.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("LLM_MODEL")

	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if modelName == "" {
		modelName = "qwen/qwen-2.5-coder-32b-instruct"
	}

	sshTimeoutSecs := 30
	if val := os.Getenv("SSH_TIMEOUT_SECONDS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			sshTimeoutSecs = parsed
		}
	}
	sshTimeout := time.Duration(sshTimeoutSecs) * time.Second

	llmTimeoutSecs := 15
	if val := os.Getenv("LLM_TIMEOUT_SECONDS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			llmTimeoutSecs = parsed
		}
	}
	llmTimeout := time.Duration(llmTimeoutSecs) * time.Second

	analyzer := llm.NewLocalOpenAIClient(baseURL, apiKey, modelName)
	
	masterSecret := []byte("a-very-secret-key-32-bytes-long!!")
	vault := security.NewLocalEncryptedVault(masterSecret)
	
	sshClient := ssh.NewSSHClient(vault)
	idempHelper := ssh.NewLinuxIdempotencyAnalyzer(sshClient)
	
	engine := agent.NewEngine(sshClient, repo, analyzer, idempHelper, taskChan, logChan, sshTimeout, llmTimeout)

	targets, err := inventory.LoadInventory("hosts.yaml")
	if err != nil {
		log.Fatalf("Failed to load inventory: %v", err)
	}

	var tasks []agent.Task
	
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

	model := tui.NewModel(taskChan, logChan, hitlChan, tasks)

	go func() {
		for i := range tasks {
			t := tasks[i]
			time.Sleep(1 * time.Second) // Stagger execution slightly
			engine.RunTask(context.Background(), t, hitlChan)
		}
	}()

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
