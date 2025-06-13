package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// MockGitConfig represents the configuration for mock git behavior
type MockGitConfig struct {
	// ExitCode is the exit code to return
	ExitCode int `json:"exit_code"`
	// Output is the stdout output to produce
	Output string `json:"output"`
	// Error is the stderr output to produce
	Error string `json:"error"`
}

func main() {
	// Read config from environment
	configJSON := os.Getenv("MOCK_GIT_CONFIG")
	if configJSON == "" {
		fmt.Fprintln(os.Stderr, "MOCK_GIT_CONFIG environment variable not set")
		os.Exit(1)
	}

	var config MockGitConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse MOCK_GIT_CONFIG: %v\n", err)
		os.Exit(1)
	}

	// Print output to stdout if any
	if config.Output != "" {
		fmt.Print(config.Output)
	}

	// Print error to stderr if any
	if config.Error != "" {
		fmt.Fprint(os.Stderr, config.Error)
	}

	// Exit with configured code
	os.Exit(config.ExitCode)
}
