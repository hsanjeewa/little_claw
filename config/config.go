package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	LLM   LLMConfig   `mapstructure:"llm"`
	Agent AgentConfig `mapstructure:"agent"`
}

type LLMConfig struct {
	BaseURL        string `mapstructure:"base_url"`
	Model          string `mapstructure:"model"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

type AgentConfig struct {
	SSHTimeoutSeconds int    `mapstructure:"ssh_timeout_seconds"`
	InventoryPath     string `mapstructure:"inventory_path"`
	DatabasePath      string `mapstructure:"database_path"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GetConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.toml"
	}
	return LoadConfig(configPath)
}