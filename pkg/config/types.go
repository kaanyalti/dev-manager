package config

import "time"

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
