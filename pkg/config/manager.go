package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager handles configuration operations
type Manager struct {
	config     *Config
	configPath string
}

// NewManager creates a new configuration manager
func NewManager(configPath string) (*Manager, error) {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(home, ".config", "dev-manager", "config.yaml")
	}

	return &Manager{
		configPath: configPath,
	}, nil
}

// Load reads the configuration file
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.config = &Config{}
			return nil
		}
		return err
	}

	m.config = &Config{}
	return yaml.Unmarshal(data, m.config)
}

// Save writes the configuration to file
func (m *Manager) Save() error {
	if m.config == nil {
		m.config = &Config{}
	}

	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	if m.config == nil {
		m.config = &Config{}
	}
	return m.config
}

// SetConfig updates the current configuration
func (m *Manager) SetConfig(cfg *Config) {
	m.config = cfg
}

// Path returns the config file path
func (m *Manager) Path() string {
	return m.configPath
}
