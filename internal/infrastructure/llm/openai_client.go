package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/devops/agent/internal/domain/agent"
)

var errMissingKeyColon = regexp.MustCompile(`"(\w+):\s*([^"]*?)"`)

type LocalOpenAIClient struct {
	client *openai.Client
	model  string
}

func NewLocalOpenAIClient(baseURL, apiKey, model string) *LocalOpenAIClient {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	return &LocalOpenAIClient{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}
}

func (c *LocalOpenAIClient) AnalyzeOutput(ctx context.Context, command string, output string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert Linux System Administrator AI.
Analyze the following shell command and its execution output.
Determine if the command succeeded, failed, or requires human intervention.
Provide a very concise summary (max 2 sentences).

Command: %s
Output:
%s`, command, output)

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.2,
		},
	)

	if err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to analyze output via LLM: %v", err))
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("context: %w", fmt.Errorf("LLM returned no response"))
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *LocalOpenAIClient) GeneratePlan(ctx context.Context, goal string, scope string, capabilities string, constraints string, watchtowerContext string) ([]PlanStep, string, error) {
	messages := BuildPlanContext(goal, scope, capabilities, constraints, watchtowerContext)

	openaiMessages := make([]openai.ChatCompletionMessage, 0, len(messages)+2)
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: msg,
		})
	}

	openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: goal,
	})

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    openaiMessages,
			Temperature: 0.2,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return nil, "", fmt.Errorf("API call failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, "", fmt.Errorf("LLM returned no response")
	}

	responseContent := resp.Choices[0].Message.Content

	responseContent = stripMarkdownFences(responseContent)
	responseContent = fixJSON(responseContent)

	reasoning := ""
	var generatedPlan []PlanStep
	if err := json.Unmarshal([]byte(responseContent), &generatedPlan); err != nil {
		var planWithReasoning GeneratedPlan
		if err2 := json.Unmarshal([]byte(responseContent), &planWithReasoning); err2 == nil && len(planWithReasoning.Steps) > 0 {
			generatedPlan = planWithReasoning.Steps
			reasoning = planWithReasoning.Reasoning
		} else {
			var singleStep PlanStep
			if err3 := json.Unmarshal([]byte(responseContent), &singleStep); err3 == nil && singleStep.Description != "" {
				generatedPlan = []PlanStep{singleStep}
			} else {
				return nil, "", fmt.Errorf("LLM response is not valid JSON plan: %v", err)
			}
		}
	}

	if len(generatedPlan) == 0 {
		return nil, "", fmt.Errorf("LLM generated empty plan")
	}

	if err := ValidatePlan(generatedPlan); err != nil {
		return nil, "", fmt.Errorf("LLM plan validation: %v", err)
	}

	return generatedPlan, reasoning, nil
}

func stripMarkdownFences(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "```") {
		if idx := strings.IndexByte(s, '\n'); idx != -1 {
			s = strings.TrimSpace(s[idx:])
		}
	}
	s = strings.TrimSuffix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func fixJSON(raw string) string {
	// Fix common LLM JSON mistakes where the colon is inside the key quotes:
	//   "command: sudo apt-get update"  →  "command": "sudo apt-get update"
	// This happens when the LLM drops the closing quote on the key and the opening quote on the value.
	result := errMissingKeyColon.ReplaceAllString(raw, `"$1": "$2"`)
	return result
}

func (c *LocalOpenAIClient) PlanTasks(ctx context.Context, goal string) ([]agent.Task, error) {
	scope := "web-prod-01, db-master, cache-01"
	capabilities := "Linux, SSH access, sudo privileges"
	constraints := "No downtime allowed, must notify before mutative operations"
	watchtowerContext := "No system state available"

	plan, _, err := c.GeneratePlan(ctx, goal, scope, capabilities, constraints, watchtowerContext)
	if err != nil {
		return nil, err
	}

	tasks := make([]agent.Task, len(plan))
	for i, step := range plan {
		task, err := agent.NewTask(
			fmt.Sprintf("task-%d", i),
			"localhost",
			"127.0.0.1",
			22,
			"root",
			step.Command,
			step.IsMutative,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create task %d: %w", i, err)
		}
		tasks[i] = task
	}

	return tasks, nil
}