package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

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

// selectKey interactively prompts the user to select a key from the list of available keys.
// Returns the selected key path or empty string if aborted.
func selectKey(action string) string {
	mgr := newSSHManager()
	keys, err := mgr.ListPrivateKeys()
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}

	if len(keys) == 0 {
		log.Fatal("no SSH keys found")
	}

	fmt.Println("Available SSH keys:")
	for i, key := range keys {
		fmt.Printf("%d. %s\n", i+1, key)
	}

	// Prompt for selection
	fmt.Printf("\nSelect a key to %s (number, or press enter to abort): ", action)
	var selectionStr string
	fmt.Scanln(&selectionStr)

	// If empty input, abort
	if selectionStr == "" {
		fmt.Println("Operation aborted.")
		return ""
	}

	// Convert selection to number
	selection, err := strconv.Atoi(selectionStr)
	if err != nil || selection < 1 || selection > len(keys) {
		log.Fatal("invalid selection")
	}

	return keys[selection-1]
}

var sshPrintPublicCmd = &cobra.Command{
	Use:   "print-public",
	Short: "Print the public key",
	Long: `Print the public key for an existing SSH private key.
If no key is specified with --key, you will be prompted to select one from a list.

Example:
  dev-manager ssh print-public --key ~/.ssh/my-key
  dev-manager ssh print-public`,
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")

		if keyPath == "" {
			keyPath = selectKey("print")
			if keyPath == "" {
				return
			}
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
If no key is specified with --key, you will be prompted to select one from a list.

Example:
  dev-manager ssh copy-public --key ~/.ssh/my-key
  dev-manager ssh copy-public`,
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")

		if keyPath == "" {
			keyPath = selectKey("copy")
			if keyPath == "" {
				return
			}
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
If no key is specified with --key, you will be prompted to select one from a list.

Example:
  dev-manager ssh remove --key ~/.ssh/my-key
  dev-manager ssh remove`,
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")

		if keyPath == "" {
			keyPath = selectKey("remove")
			if keyPath == "" {
				return
			}
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
		}
		fmt.Printf("Removed public key: %s\n", pubKeyPath)
	},
}

var sshListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available SSH key pairs and agent-loaded keys",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := newSSHManager()

		fmt.Println("Private SSH keys in ~/.ssh:")
		keys, err := mgr.ListPrivateKeys()
		if err != nil {
			log.Fatalf("Failed to list SSH keys: %v", err)
		}
		if len(keys) == 0 {
			fmt.Println("  (none found)")
		} else {
			for _, k := range keys {
				fmt.Println("  ", k)
			}
		}

		fmt.Println("\nKeys loaded in ssh-agent:")
		agentKeys, err := mgr.ListAgentKeys()
		if err != nil {
			log.Fatalf("Failed to list agent keys: %v", err)
		}
		if len(agentKeys) == 0 {
			fmt.Println("  (none loaded)")
		} else {
			for _, k := range agentKeys {
				fmt.Println("  ", k)
			}
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

	sshCmd.AddCommand(sshListCmd)
}
