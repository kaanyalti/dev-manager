package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Repository handles git operations for a single repository
type Repository struct {
	Path   string
	URL    string
	Branch string
}

// New creates a new Repository instance
func New(path, url, branch string) *Repository {
	if branch == "" {
		branch = "main"
	}
	return &Repository{
		Path:   path,
		URL:    url,
		Branch: branch,
	}
}

// Clone clones the repository if it doesn't exist
func (r *Repository) Clone() error {
	if _, err := os.Stat(r.Path); !os.IsNotExist(err) {
		return fmt.Errorf("path already exists: %s", r.Path)
	}

	if err := os.MkdirAll(filepath.Dir(r.Path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	cmd := exec.Command("git", "clone", "-b", r.Branch, r.URL, r.Path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// Update fetches and rebases the repository
func (r *Repository) Update() error {
	// Check if directory exists
	if _, err := os.Stat(r.Path); os.IsNotExist(err) {
		return r.Clone()
	}

	// Fetch updates
	fetchCmd := exec.Command("git", "-C", r.Path, "fetch", "origin", r.Branch)
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch updates: %s, %w", string(output), err)
	}

	// Rebase
	rebaseCmd := exec.Command("git", "-C", r.Path, "rebase", fmt.Sprintf("origin/%s", r.Branch))
	if output, err := rebaseCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to rebase: %s, %w", string(output), err)
	}

	return nil
}

// IsClean checks if the repository has any uncommitted changes
func (r *Repository) IsClean() (bool, error) {
	cmd := exec.Command("git", "-C", r.Path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check repository status: %w", err)
	}

	return len(output) == 0, nil
}
