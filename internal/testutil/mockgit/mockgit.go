package mockgit

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// MockGit represents a mock git binary
type MockGit struct {
	// Path is the path to the mock git binary
	Path string
	// OriginalPath is the original PATH value
	OriginalPath string
}

// Config represents the configuration for mock git behavior
type Config struct {
	// ExitCode is the exit code to return
	ExitCode int
	// Output is the stdout output to produce
	Output string
	// Error is the stderr output to produce
	Error string
}

// New creates a new mock git binary for testing
func New(t *testing.T) *MockGit {
	t.Helper()

	// Skip on Windows as PATH manipulation is different
	if runtime.GOOS == "windows" {
		t.Skip("Mock git tests are not supported on Windows")
	}

	// Create temp directory for the mock binary
	tempDir := t.TempDir()
	mockPath := filepath.Join(tempDir, "git")

	// Build the mock git binary
	cmd := exec.Command("go", "build", "-o", mockPath, "github.com/kaanyalti/dev-manager/internal/testutil/mockgit/cmd")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build mock git: %v", err)
	}

	// Save original PATH
	originalPath := os.Getenv("PATH")

	// Set PATH to use our mock git
	newPath := tempDir + string(os.PathListSeparator) + originalPath
	if err := os.Setenv("PATH", newPath); err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	return &MockGit{
		Path:         mockPath,
		OriginalPath: originalPath,
	}
}

// Configure sets the behavior of the mock git
func (m *MockGit) Configure(t *testing.T, config Config) {
	t.Helper()

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.Setenv("MOCK_GIT_CONFIG", string(configJSON)); err != nil {
		t.Fatalf("Failed to set MOCK_GIT_CONFIG: %v", err)
	}
}

// Cleanup restores the original PATH
func (m *MockGit) Cleanup() {
	os.Setenv("PATH", m.OriginalPath)
	os.Unsetenv("MOCK_GIT_CONFIG")
}
