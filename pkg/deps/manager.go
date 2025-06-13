package deps

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dev-manager/pkg/config"
)

// Manager handles dependency operations
type Manager struct {
	InstallDir string
}

// New creates a new dependency manager
func New(installDir string) *Manager {
	return &Manager{
		InstallDir: installDir,
	}
}

// Install installs a dependency
func (m *Manager) Install(dep config.Dependency, force bool) error {
	// Create installation directory if it doesn't exist
	if err := os.MkdirAll(m.InstallDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Check if already installed
	depPath := filepath.Join(m.InstallDir, dep.Name)
	if _, err := os.Stat(depPath); err == nil && !force {
		return fmt.Errorf("%s is already installed at %s", dep.Name, depPath)
	}

	// Download the dependency
	resp, err := http.Get(dep.Source)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", dep.Name, err)
	}
	defer resp.Body.Close()

	// Create temporary directory for extraction
	tmpDir, err := os.MkdirTemp("", "dev-manager-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Handle different file types
	switch {
	case strings.HasSuffix(dep.Source, ".tar.gz"):
		if err := extractTarGz(resp.Body, tmpDir); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	case strings.HasSuffix(dep.Source, ".zip"):
		// TODO: Implement zip extraction
		return fmt.Errorf("zip extraction not implemented yet")
	default:
		// Assume it's a binary, just copy it
		out, err := os.Create(filepath.Join(tmpDir, dep.Name))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, resp.Body); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	// Move to final location
	if err := os.RemoveAll(depPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing installation: %w", err)
	}

	if err := os.Rename(tmpDir, depPath); err != nil {
		return fmt.Errorf("failed to move to final location: %w", err)
	}

	// Make executable if it's a binary
	if err := makeExecutable(depPath); err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	return nil
}

// Remove removes a dependency
func (m *Manager) Remove(dep config.Dependency) error {
	depPath := filepath.Join(m.InstallDir, dep.Name)
	if err := os.RemoveAll(depPath); err != nil {
		return fmt.Errorf("failed to remove %s: %w", dep.Name, err)
	}
	return nil
}

// Helper functions

func extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func makeExecutable(path string) error {
	// If it's a directory, find the main binary
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		// Look for common binary names
		binaryNames := []string{"bin", "sbin", "exec", "main"}
		for _, name := range binaryNames {
			binaryPath := filepath.Join(path, name)
			if _, err := os.Stat(binaryPath); err == nil {
				path = binaryPath
				break
			}
		}
	}

	// Make executable
	return os.Chmod(path, 0755)
}
