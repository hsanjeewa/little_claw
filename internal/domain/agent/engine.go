package agent

import (
	"context"
	"fmt"
	"time"
)

type CommandExecutor interface {
	Execute(ctx context.Context, task Task) (string, error)
}

type Engine struct {
	executor  CommandExecutor
	repo      AuditRepository
	analyzer  AIAnalyzer
	idemp     IdempotencyHelper
	taskChan  chan<- Task
	logChan   chan<- ExecutionLog
}

func NewEngine(
	executor CommandExecutor,
	repo AuditRepository,
	analyzer AIAnalyzer,
	idemp IdempotencyHelper,
	taskChan chan<- Task,
	logChan chan<- ExecutionLog,
) *Engine {
	return &Engine{
		executor:  executor,
		repo:      repo,
		analyzer:  analyzer,
		idemp:     idemp,
		taskChan:  taskChan,
		logChan:   logChan,
	}
}

func (e *Engine) RunTask(ctx context.Context, task Task, hitlChan chan<- HitlRequest) {
	isIdemp, msg := e.idemp.IsSatisfied(ctx, task)
	if isIdemp {
		task.Status = StatusIdempotent
		e.taskChan <- task
		e.logAndNotify(ctx, task, msg, "Task skipped (Idempotent)")
		return
	}

	if task.IsMutative {
		task.Status = "WAITING"
		e.taskChan <- task

		respChan := make(chan bool)
		hitlChan <- HitlRequest{Task: task, ResponseChan: respChan}
		
		select {
		case <-ctx.Done():
			task.Status = StatusFailed
			e.taskChan <- task
			e.logAndNotify(ctx, task, "Timeout/Cancel waiting for HITL", "Aborted")
			return
		case approved := <-respChan:
			if !approved {
				task.Status = StatusSkipped
				e.taskChan <- task
				e.logAndNotify(ctx, task, "Operator Denied Authorization", "Skipped")
				return
			}
		}
	}

	task.Status = StatusRunning
	e.taskChan <- task

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	output, err := e.executor.Execute(execCtx, task)
	
	if err != nil {
		task.Status = StatusFailed
	} else if task.IsMutative {
		task.Status = StatusChanged
	} else {
		task.Status = StatusSuccess
	}
	e.taskChan <- task

	aiCtx, aiCancel := context.WithTimeout(ctx, 15*time.Second)
	defer aiCancel()
	
	aiAnalysis, aiErr := e.analyzer.AnalyzeOutput(aiCtx, task.Command, output)
	if aiErr != nil {
		aiAnalysis = fmt.Sprintf("AI Analysis Failed (Network or Timeout): %v", aiErr)
	}

	finalOutput := fmt.Sprintf("RAW OUTPUT:\n%s\n\n[AI ANALYSIS]\n%s", output, aiAnalysis)
	e.logAndNotify(ctx, task, finalOutput, "Completed")
}

func (e *Engine) logAndNotify(ctx context.Context, task Task, output string, summary string) {
	execLog := ExecutionLog{
		Timestamp: time.Now(),
		Host:      task.HostAlias,
		Command:   task.Command,
		Status:    task.Status,
		Output:    output,
	}
	
	if err := e.repo.SaveLog(ctx, execLog); err != nil {
		_ = err 
	}
	e.logChan <- execLog
}
