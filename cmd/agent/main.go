package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/database"
	"github.com/devops/agent/internal/infrastructure/llm"
	"github.com/devops/agent/internal/ui/tui"
)

func main() {
	_ = godotenv.Load()

	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan tui.HitlRequest, 10)

	dbPath := "./agent.db"
	repo, err := database.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	task1, _ := agent.NewTask(uuid.New().String(), "local-web", "127.0.0.1", 22, "root", "uptime", false)
	task2, _ := agent.NewTask(uuid.New().String(), "local-db", "127.0.0.1", 22, "root", "df -h", false)
	task3, _ := agent.NewTask(uuid.New().String(), "local-cache", "127.0.0.1", 22, "root", "sudo whoami", true)

	tasks := []agent.Task{task1, task2, task3}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("LLM_MODEL")

	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if modelName == "" {
		modelName = "qwen/qwen-2.5-coder-32b-instruct"
	}

	analyzer := llm.NewLocalOpenAIClient(baseURL, apiKey, modelName)

	model := tui.NewModel(taskChan, logChan, hitlChan, tasks)

	go simulateExecution(tasks, repo, analyzer, taskChan, logChan, hitlChan)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func simulateExecution(tasks []agent.Task, repo agent.AuditRepository, analyzer agent.AIAnalyzer, taskChan chan agent.Task, logChan chan agent.ExecutionLog, hitlChan chan tui.HitlRequest) {
	for {
		for i := range tasks {
			t := tasks[i]
			
			time.Sleep(2 * time.Second)
			
			if t.IsMutative {
				t.Status = "WAITING"
				taskChan <- t

				respChan := make(chan bool)
				hitlChan <- tui.HitlRequest{Task: t, ResponseChan: respChan}
				approved := <-respChan

				if !approved {
					t.Status = agent.StatusSkipped
					taskChan <- t
					
					execLog := agent.ExecutionLog{
						ID:        uuid.New().String(),
						Timestamp: time.Now(),
						Host:      t.HostIP,
						Command:   t.Command,
						Status:    t.Status,
						Output:    "Operator Denied Authorization",
					}
					_ = repo.SaveLog(context.Background(), execLog)
					logChan <- execLog
					continue
				}
			}

			t.Status = agent.StatusRunning
			taskChan <- t
			
			time.Sleep(1 * time.Second)
			
			if rand.Intn(2) == 0 {
				t.Status = agent.StatusSuccess
			} else {
				t.Status = agent.StatusFailed
			}
			taskChan <- t
			
			rawOutput := fmt.Sprintf("Simulated output for %s", t.Command)
			aiAnalysis, err := analyzer.AnalyzeOutput(context.Background(), t.Command, rawOutput)
			if err != nil {
				aiAnalysis = fmt.Sprintf("AI Analysis Failed: %v", err)
			}
			
			execLog := agent.ExecutionLog{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Host:      t.HostIP,
				Command:   t.Command,
				Status:    t.Status,
				Output:    fmt.Sprintf("%s\n\n[AI ANALYSIS]\n%s", rawOutput, aiAnalysis),
			}
			
				if err := repo.SaveLog(context.Background(), execLog); err != nil {
					_ = err
				}
			
			logChan <- execLog
		}
	}
}
