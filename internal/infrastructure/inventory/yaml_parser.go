package inventory

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type HostConfig struct {
	AnsibleHost string `yaml:"ansible_host"`
	AnsiblePort int    `yaml:"ansible_port"`
	AnsibleUser string `yaml:"ansible_user"`
}

type GroupConfig struct {
	Hosts map[string]HostConfig `yaml:"hosts"`
}

type AnsibleInventory struct {
	All struct {
		Groups map[string]GroupConfig `yaml:"groups"`
	} `yaml:"all"`
}

type TargetHost struct {
	Alias string
	IP    string
	Port  int
	User  string
}

func LoadInventory(filePath string) ([]TargetHost, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to read inventory file: %v", err))
	}

	var inv AnsibleInventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("failed to parse yaml: %v", err))
	}

	var targets []TargetHost
	for _, group := range inv.All.Groups {
		for alias, config := range group.Hosts {
			
			port := config.AnsiblePort
			if port == 0 {
				port = 22
			}
			
			user := config.AnsibleUser
			if user == "" {
				user = "root"
			}
			
			ip := config.AnsibleHost
			if ip == "" {
				ip = alias
			}
			
			targets = append(targets, TargetHost{
				Alias: alias,
				IP:    ip,
				Port:  port,
				User:  user,
			})
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("inventory is empty or improperly formatted"))
	}

	return targets, nil
}
