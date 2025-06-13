package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"dev-manager/internal/ssh"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// newSSHManager is a helper to create a new SSHManager and handle errors.
func newSSHManager() *ssh.SSHManager {
	mgr, err := ssh.NewSSHManager()
	if err != nil {
		log.Fatalf("Failed to initialize SSH manager: %v", err)
	}
	return mgr
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Manage SSH keys",
	Long:  `Commands for managing SSH keys.`,
}

var sshGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new SSH key",
	Long: `Generate a new SSH key with the specified algorithm and name.
Supported algorithms: rsa, ed25519.

Example:
  dev-manager ssh generate --algo ed25519 --name my-key
  dev-manager ssh generate -a rsa -n another-key`,
	Run: func(cmd *cobra.Command, args []string) {
		algo, _ := cmd.Flags().GetString("algo")
		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			log.Fatal("key name is required (--name)")
		}

		mgr := newSSHManager()
		keyPath, err := mgr.GenerateKey(algo, name)
		if err != nil {
			log.Fatalf("failed to generate key: %v", err)
		}

		fmt.Printf("Generated SSH key: %s\n", keyPath)
	},
}

var sshAddAgentCmd = &cobra.Command{
	Use:   "add-agent",
	Short: "Add a key to SSH agent",
	Long: `Add an existing SSH key to the SSH agent.
The key must be unencrypted.

Example:
  dev-manager ssh add-agent --key ~/.ssh/my-key`,
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")

		if keyPath == "" {
			log.Fatal("key path is required (--key)")
		}

		mgr := newSSHManager()
		if err := mgr.AddKeyToAgent(keyPath); err != nil {
			log.Fatalf("failed to add key to agent: %v", err)
		}

		fmt.Printf("Added key to SSH agent: %s\n", keyPath)
	},
}

var sshPrintPublicCmd = &cobra.Command{
	Use:   "print-public",
	Short: "Print the public key",
	Long: `Print the public key for an existing SSH private key.
Example:
  dev-manager ssh print-public --key ~/.ssh/my-key`,
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")

		if keyPath == "" {
			log.Fatal("key path is required (--key)")
		}

		mgr := newSSHManager()
		if err := mgr.PrintPublicKey(keyPath); err != nil {
			log.Fatalf("failed to print public key: %v", err)
		}
	},
}

var sshCopyPublicCmd = &cobra.Command{
	Use:   "copy-public",
	Short: "Copy public key to clipboard",
	Long: `Copy the public key to the clipboard for an existing SSH private key.
Example:
  dev-manager ssh copy-public --key ~/.ssh/my-key`,
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")

		if keyPath == "" {
			log.Fatal("key path is required (--key)")
		}

		pubKeyPath := keyPath + ".pub"
		pubKey, err := os.ReadFile(pubKeyPath)
		if err != nil {
			log.Fatalf("failed to get public key: %v", err)
		}

		if err := clipboard.WriteAll(string(pubKey)); err != nil {
			log.Fatalf("failed to copy to clipboard: %v", err)
		}

		fmt.Println("Public key copied to clipboard.")
	},
}

var sshRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an SSH key",
	Long: `Remove an SSH key from the filesystem and agent.
Example:
  dev-manager ssh remove --key ~/.ssh/my-key`,
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")

		if keyPath == "" {
			log.Fatal("key path is required (--key)")
		}

		// Remove from agent first (best effort, ignore error if not loaded)
		_ = exec.Command("ssh-add", "-d", keyPath).Run()

		// Delete private key
		if err := os.Remove(keyPath); err != nil {
			log.Fatalf("failed to remove private key: %v", err)
		}
		fmt.Printf("Removed private key: %s\n", keyPath)

		// Delete public key
		pubKeyPath := keyPath + ".pub"
		if err := os.Remove(pubKeyPath); err != nil {
			// Don't fail if the public key doesn't exist for some reason
			if !os.IsNotExist(err) {
				log.Printf("failed to remove public key: %v\n", err)
			}
		} else {
			fmt.Printf("Removed public key: %s\n", pubKeyPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.AddCommand(sshGenerateCmd)
	sshGenerateCmd.Flags().StringP("algo", "a", "ed25519", "Key generation algorithm (rsa, ed25519)")
	sshGenerateCmd.Flags().StringP("name", "n", "", "Name of the key")

	sshCmd.AddCommand(sshAddAgentCmd)
	sshAddAgentCmd.Flags().StringP("key", "k", "", "Path to the private key")

	sshCmd.AddCommand(sshPrintPublicCmd)
	sshPrintPublicCmd.Flags().StringP("key", "k", "", "Path to the private key")

	sshCmd.AddCommand(sshCopyPublicCmd)
	sshCopyPublicCmd.Flags().StringP("key", "k", "", "Path to the private key")

	sshCmd.AddCommand(sshRemoveCmd)
	sshRemoveCmd.Flags().StringP("key", "k", "", "Path to the private key")
}
