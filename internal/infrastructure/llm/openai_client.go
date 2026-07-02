package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/devops/agent/internal/domain/agent"
)

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

func (c *LocalOpenAIClient) PlanTasks(ctx context.Context, goal string) ([]agent.Task, error) {
	return nil, fmt.Errorf("context: %w", fmt.Errorf("PlanTasks not yet fully implemented"))
}

var _ agent.AIAnalyzer = (*LocalOpenAIClient)(nil)
