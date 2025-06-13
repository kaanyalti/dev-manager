package git

import (
	"os"
	"path/filepath"
	"testing"

	"dev-manager/internal/testutil/mockgit"
)

func TestRepository_Clone(t *testing.T) {
	// Setup mock git
	mock := mockgit.New(t)
	defer mock.Cleanup()

	// Create temp directory for test
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		repo    *Repository
		config  mockgit.Config
		wantErr bool
	}{
		{
			name: "successful clone",
			repo: New(filepath.Join(tempDir, "repo"), "https://github.com/test/repo", "main"),
			config: mockgit.Config{
				ExitCode: 0,
				Output:   "Cloning into 'repo'...\n",
			},
			wantErr: false,
		},
		{
			name: "git command fails",
			repo: New(filepath.Join(tempDir, "repo"), "https://github.com/test/repo", "main"),
			config: mockgit.Config{
				ExitCode: 1,
				Error:    "fatal: repository 'https://github.com/test/repo' not found\n",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure mock git behavior
			mock.Configure(t, tt.config)

			// Run the test
			err := tt.repo.Clone()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.Clone() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check if directory exists for successful clone
			if !tt.wantErr {
				if _, err := os.Stat(tt.repo.Path); os.IsNotExist(err) {
					t.Error("Repository.Clone() did not create directory")
				}
			}
		})
	}
}
