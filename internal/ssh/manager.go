package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type SSHManager struct {
	HomeDir string
}

func NewSSHManager() (*SSHManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &SSHManager{HomeDir: home}, nil
}

// Check if required SSH tools are installed
func (m *SSHManager) CheckTools() error {
	tools := []string{"ssh", "ssh-keygen", "ssh-agent"}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found in PATH", tool)
		}
	}
	return nil
}

// Check if ssh-agent is running (by checking SSH_AUTH_SOCK)
func (m *SSHManager) IsAgentRunning() bool {
	return os.Getenv("SSH_AUTH_SOCK") != ""
}

// List private keys in ~/.ssh
func (m *SSHManager) ListPrivateKeys() ([]string, error) {
	sshDir := filepath.Join(m.HomeDir, ".ssh")
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if name == "id_rsa" || name == "id_ed25519" || name == "id_ecdsa" || name == "id_dsa" {
			keys = append(keys, filepath.Join(sshDir, name))
		}
	}
	return keys, nil
}

// List keys loaded in the agent
func (m *SSHManager) ListAgentKeys() ([]string, error) {
	cmd := exec.Command("ssh-add", "-l")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No identities loaded
			return []string{}, nil
		}
		return nil, fmt.Errorf("ssh-add -l failed: %s", string(output))
	}
	return []string{string(output)}, nil
}

// Add a key to the agent
func (m *SSHManager) AddKeyToAgent(keyPath string) error {
	cmd := exec.Command("ssh-add", keyPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Generate a new SSH key pair
func (m *SSHManager) GenerateKey(algo, name string) (string, error) {
	sshDir := filepath.Join(m.HomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return "", err
	}
	keyFile := "id_" + algo
	if name != "" {
		keyFile = name + "_id_" + algo
	}
	keyPath := filepath.Join(sshDir, keyFile)
	cmd := exec.Command("ssh-keygen", "-t", algo, "-f", keyPath, "-N", "")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return keyPath, nil
}

// Print public key and instructions
func (m *SSHManager) PrintPublicKey(keyPath string) error {
	pubPath := keyPath + ".pub"
	data, err := os.ReadFile(pubPath)
	if err != nil {
		return err
	}
	fmt.Printf("\nYour public key (%s):\n%s\n", pubPath, string(data))
	fmt.Println("\nCopy this key and add it to your GitHub/GitLab/Bitbucket account.")
	return nil
}
