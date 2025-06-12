package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"dev-manager/internal/ssh"
	"dev-manager/pkg/config"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dev-manager",
	Short: "Dev Manager - A tool to manage development environment",
	Long: `Dev Manager helps you manage your development environment by:
- Managing git repositories
- Syncing tool configurations (nvim, tmux, zsh)
- Keeping repositories up to date`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dev-manager configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("config")
		workspace, _ := cmd.Flags().GetString("workspace")

		// Default workspace: $HOME/dev
		if workspace == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("failed to get home directory: %v", err)
			}
			workspace = filepath.Join(home, "dev")
		}

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		// Attempt to load existing config (fail if parsing error, ignore if not exists)
		if err := mgr.Load(); err != nil {
			if !os.IsNotExist(err) {
				log.Fatalf("failed to load config: %v", err)
			}
		}

		cfg := mgr.GetConfig()
		if err := cfg.Validate(); err != nil {
			log.Fatalf("invalid configuration: %v", err)
		}
		if cfg.WorkspacePath == "" {
			cfg.WorkspacePath = workspace
		}
		if cfg.UpdateFrequency == 0 {
			cfg.UpdateFrequency = 2 * time.Hour
		}

		// Save configuration
		if err := mgr.Save(); err != nil {
			log.Fatalf("failed to save configuration: %v", err)
		}

		fmt.Printf("Configuration initialized at %s\n", mgr.Path())
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all repositories and configurations",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement sync
		fmt.Println("Syncing repositories and configurations...")
	},
}

var addRepoCmd = &cobra.Command{
	Use:   "add-repo [name] [url]",
	Short: "Add a repository to manage",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement add repository
		fmt.Printf("Adding repository %s from %s\n", args[0], args[1])
	},
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Manage SSH keys and agent for your dev environment",
	Long: `SSH key and agent management for dev-manager.

Planned subcommands:
  init           # Comprehensive SSH environment setup (checks, agent, keys)
  generate       # Generate a new SSH key pair
  list           # List available SSH key pairs in ~/.ssh
  add-agent      # Add a key to the ssh-agent
  status         # Show SSH environment status (tooling, agent, keys, agent keys)
  print-public   # Print the public key (for copy-paste to GitHub/GitLab)
  copy-public    # Copy the public key to clipboard (optional, cross-platform)
`,
}

var sshInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Comprehensive SSH environment setup (checks, agent, keys)",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

		fmt.Println("Checking for required SSH tools...")
		if err := mgr.CheckTools(); err != nil {
			log.Fatalf("SSH tooling check failed: %v", err)
		}
		fmt.Println("All required SSH tools are installed.")

		fmt.Println("Checking if ssh-agent is running...")
		if !mgr.IsAgentRunning() {
			fmt.Println("ssh-agent is not running. Please start ssh-agent and try again.")
			os.Exit(1)
		}
		fmt.Println("ssh-agent is running.")

		fmt.Println("Looking for existing SSH key pairs...")
		keys, err := mgr.ListPrivateKeys()
		if err != nil {
			log.Fatalf("Failed to list SSH keys: %v", err)
		}
		if len(keys) == 0 {
			fmt.Println("No SSH key pairs found. Generating a new ed25519 key pair...")
			keyPath, err := mgr.GenerateKey("ed25519", "")
			if err != nil {
				log.Fatalf("Failed to generate SSH key: %v", err)
			}
			keys = append(keys, keyPath)
			fmt.Println("New SSH key pair generated.")
		} else {
			fmt.Printf("Found %d SSH key(s):\n", len(keys))
			for _, k := range keys {
				fmt.Println("  ", k)
			}
		}

		fmt.Println("Checking if any SSH key is loaded in the agent...")
		agentKeys, err := mgr.ListAgentKeys()
		if err != nil {
			log.Fatalf("Failed to list agent keys: %v", err)
		}
		if len(agentKeys) == 0 {
			fmt.Println("No SSH keys loaded in the agent. Adding the first available key...")
			if err := mgr.AddKeyToAgent(keys[0]); err != nil {
				log.Fatalf("Failed to add key to agent: %v", err)
			}
			fmt.Println("Key added to agent.")
		} else {
			fmt.Println("SSH key(s) already loaded in the agent.")
		}

		// Print public key and instructions
		if err := mgr.PrintPublicKey(keys[0]); err != nil {
			log.Printf("Could not print public key: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(addRepoCmd)

	// Flags for init command
	initCmd.Flags().StringP("config", "c", "", "Path to configuration file")
	initCmd.Flags().StringP("workspace", "w", "", "Workspace directory for cloning repositories")

	// Flags for add-repo command
	addRepoCmd.Flags().StringP("branch", "b", "main", "Branch to track")
	addRepoCmd.Flags().StringP("path", "p", "", "Custom path for the repository")

	// SSH command group and subcommands
	sshCmd.AddCommand(sshInitCmd)
	rootCmd.AddCommand(sshCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
