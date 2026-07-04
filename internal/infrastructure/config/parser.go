package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	LLM struct {
		BaseURL        string `toml:"base_url"`
		Model          string `toml:"model"`
		TimeoutSeconds int    `toml:"timeout_seconds"`
	} `toml:"llm"`
	Agent struct {
		SSHTimeoutSeconds int    `toml:"ssh_timeout_seconds"`
		InventoryPath     string `toml:"inventory_path"`
		DatabasePath      string `toml:"database_path"`
		WatchtowerBackend string `toml:"watchtower_backend"`
	} `toml:"agent"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return getDefaultConfig(), nil
		}
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to read config file: %v", err))
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to parse config file: %v", err))
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func getDefaultConfig() *Config {
	var cfg Config
	applyDefaults(&cfg)
	return &cfg
}

func applyDefaults(cfg *Config) {
	if cfg.LLM.BaseURL == "" {
		cfg.LLM.BaseURL = "https://openrouter.ai/api/v1"
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "qwen/qwen-2.5-coder-32b-instruct"
	}
	if cfg.LLM.TimeoutSeconds == 0 {
		cfg.LLM.TimeoutSeconds = 15
	}
	if cfg.Agent.SSHTimeoutSeconds == 0 {
		cfg.Agent.SSHTimeoutSeconds = 30
	}
	if cfg.Agent.InventoryPath == "" {
		cfg.Agent.InventoryPath = "hosts.yaml"
	}
	if cfg.Agent.DatabasePath == "" {
		cfg.Agent.DatabasePath = "./agent.db"
	}
	if cfg.Agent.WatchtowerBackend == "" {
		cfg.Agent.WatchtowerBackend = "ssh"
	}
}
