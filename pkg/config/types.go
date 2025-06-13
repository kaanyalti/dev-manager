package config

import (
	"fmt"
	"strings"
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

// Dependency represents a development dependency
type Dependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Source  string `yaml:"source"` // URL or source location
	Path    string `yaml:"path"`   // Installation path
}

// Config represents the main configuration structure
type Config struct {
	Repositories    []Repository  `yaml:"repositories"`
	Tools           []ToolConfig  `yaml:"tools"`
	Dependencies    []Dependency  `yaml:"dependencies"`
	UpdateFrequency time.Duration `yaml:"updateFrequency"`
	WorkspacePath   string        `yaml:"workspacePath"`
}

// ValidationError represents a collection of configuration validation errors
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}
	report := "Configuration validation failed:\n"
	for _, err := range e.Errors {
		report += fmt.Sprintf("  - %s\n", err)
	}
	return report
}

// Validate checks the configuration for required fields and structure
func (c *Config) Validate() error {
	var errors []string

	// Validate workspace path
	if c.WorkspacePath == "" {
		errors = append(errors, "workspacePath is required")
	}

	// Validate update frequency
	if c.UpdateFrequency <= 0 {
		errors = append(errors, "updateFrequency must be positive")
	}

	// Validate repositories
	for i, repo := range c.Repositories {
		repoErrors := []string{}
		if repo.Name == "" {
			repoErrors = append(repoErrors, "missing name")
		}
		if repo.URL == "" {
			repoErrors = append(repoErrors, "missing url")
		}
		if repo.Path == "" {
			repoErrors = append(repoErrors, "missing path")
		}
		if repo.Branch == "" {
			repoErrors = append(repoErrors, "missing branch")
		}
		if len(repoErrors) > 0 {
			errors = append(errors, fmt.Sprintf("repository[%d] (%s): %s", i, repo.Name, strings.Join(repoErrors, ", ")))
		}
	}

	// Validate tools
	for i, tool := range c.Tools {
		toolErrors := []string{}
		if tool.Name == "" {
			toolErrors = append(toolErrors, "missing name")
		}
		if tool.ConfigPath == "" {
			toolErrors = append(toolErrors, "missing configPath")
		}
		if len(toolErrors) > 0 {
			errors = append(errors, fmt.Sprintf("tool[%d] (%s): %s", i, tool.Name, strings.Join(toolErrors, ", ")))
		}
	}

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}
	return nil
}
