package config

import (
	"fmt"
	"time"
)

// Repository represents a Git repository to be managed
type Repository struct {
	Name     string    `yaml:"name"`
	URL      string    `yaml:"url"`
	Branch   string    `yaml:"branch"`
	Path     string    `yaml:"path"`
	LastSync time.Time `yaml:"lastSync"`
}

// ToolConfig represents configuration for development tools
type ToolConfig struct {
	Name       string `yaml:"name"`
	ConfigPath string `yaml:"configPath"`
	BackupPath string `yaml:"backupPath"`
}

// Config represents the main configuration structure
type Config struct {
	Repositories    []Repository  `yaml:"repositories"`
	Tools           []ToolConfig  `yaml:"tools"`
	UpdateFrequency time.Duration `yaml:"updateFrequency"`
	WorkspacePath   string        `yaml:"workspacePath"`
}

// Validate checks the configuration for required fields and structure
func (c *Config) Validate() error {
	if c.WorkspacePath == "" {
		return fmt.Errorf("workspacePath is required")
	}
	if c.UpdateFrequency <= 0 {
		return fmt.Errorf("updateFrequency must be positive")
	}
	for i, repo := range c.Repositories {
		if repo.Name == "" {
			return fmt.Errorf("repository[%d] is missing name", i)
		}
		if repo.URL == "" {
			return fmt.Errorf("repository[%d] is missing url", i)
		}
		if repo.Path == "" {
			return fmt.Errorf("repository[%d] is missing path", i)
		}
	}
	for i, tool := range c.Tools {
		if tool.Name == "" {
			return fmt.Errorf("tool[%d] is missing name", i)
		}
		if tool.ConfigPath == "" {
			return fmt.Errorf("tool[%d] is missing configPath", i)
		}
	}
	return nil
}
