package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
