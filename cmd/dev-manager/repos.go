package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"dev-manager/pkg/config"
	"dev-manager/pkg/git"

	"github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "Manage repositories",
	Long:  `Commands for managing repositories in your workspace.`,
}

var repoAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a repository to manage",
	Long: `Add a new repository to be managed by dev-manager.
The repository will be cloned to the workspace directory under the specified name.

Example:
  dev-manager repos add --name my-project --url https://github.com/username/my-project.git`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no flags are provided
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("url") {
			cmd.Help()
			os.Exit(0)
		}

		cfgPath, _ := cmd.Flags().GetString("file")
		repoName, _ := cmd.Flags().GetString("name")
		repoURL, _ := cmd.Flags().GetString("url")

		if repoName == "" {
			log.Fatal("repository name is required (--name)")
		}
		if repoURL == "" {
			log.Fatal("repository URL is required (--url)")
		}

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		if err := mgr.Load(); err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		cfg := mgr.GetConfig()

		// Check if repository already exists
		for _, repo := range cfg.Repositories {
			if repo.Name == repoName {
				log.Fatalf("repository with name '%s' already exists", repoName)
			}
		}

		// Create repository path
		repoPath := filepath.Join(cfg.WorkspacePath, repoName)

		// Add new repository
		newRepo := config.Repository{
			Name:     repoName,
			URL:      repoURL,
			Path:     repoPath,
			Branch:   "main", // Default to main branch
			LastSync: time.Now(),
		}

		cfg.Repositories = append(cfg.Repositories, newRepo)

		// Save configuration
		if err := mgr.Save(); err != nil {
			log.Fatalf("failed to save configuration: %v", err)
		}

		fmt.Printf("Added repository '%s' from %s\n", repoName, repoURL)
		fmt.Printf("Repository will be cloned to: %s\n", repoPath)

		// Prompt for immediate cloning
		fmt.Print("Would you like to clone the repository now? (Y/n): ")
		var resp string
		fmt.Scanln(&resp)
		if resp == "" || resp == "Y" || resp == "y" {
			fmt.Println("Cloning repository...")
			repo := git.New(newRepo.Path, newRepo.URL, newRepo.Branch)
			if err := repo.Clone(); err != nil {
				log.Fatalf("failed to clone repository: %v", err)
			}
			fmt.Println("Repository cloned successfully.")
		}
	},
}

var repoRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a managed repository",
	Long: `Remove a repository from dev-manager's configuration.
Does not delete the repository from the filesystem.

Example:
  dev-manager repos remove --name my-project`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("file")
		repoName, _ := cmd.Flags().GetString("name")

		if repoName == "" {
			log.Fatal("repository name is required (--name)")
		}

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		if err := mgr.Load(); err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		cfg := mgr.GetConfig()

		found := false
		for i, repo := range cfg.Repositories {
			if repo.Name == repoName {
				cfg.Repositories = append(cfg.Repositories[:i], cfg.Repositories[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("repository with name '%s' not found", repoName)
		}

		// Save configuration
		if err := mgr.Save(); err != nil {
			log.Fatalf("failed to save configuration: %v", err)
		}

		fmt.Printf("Removed repository '%s' from management.\n", repoName)
	},
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed repositories",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("file")

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		if err := mgr.Load(); err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		cfg := mgr.GetConfig()

		if len(cfg.Repositories) == 0 {
			fmt.Println("No repositories configured.")
			return
		}

		fmt.Printf("Managed repositories (%d):\n\n", len(cfg.Repositories))
		for _, repo := range cfg.Repositories {
			fmt.Printf("Name: %s\n", repo.Name)
			fmt.Printf("  URL: %s\n", repo.URL)
			fmt.Printf("  Path: %s\n", repo.Path)
			fmt.Printf("  Branch: %s\n", repo.Branch)
			fmt.Printf("  Last Sync: %s\n", repo.LastSync.Format(time.RFC3339))
			fmt.Println()
		}
	},
}

var repoSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync a specific repository",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement sync for a specific repository
		fmt.Println("Syncing a specific repository...")
	},
}

var repoSyncAllCmd = &cobra.Command{
	Use:   "sync-all",
	Short: "Sync all repositories",
	Long:  `Sync all repositories by pulling the latest changes from their remotes.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("file")

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		if err := mgr.Load(); err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		cfg := mgr.GetConfig()

		for _, repo := range cfg.Repositories {
			fmt.Printf("Syncing repository: %s...\n", repo.Name)
			r := git.New(repo.Path, repo.URL, repo.Branch)
			if err := r.Update(); err != nil {
				log.Printf("failed to sync repository %s: %v\n", repo.Name, err)
				continue
			}
			fmt.Printf("Synced repository: %s\n", repo.Name)
		}
	},
}

func init() {
	// Add repo commands
	rootCmd.AddCommand(reposCmd)
	reposCmd.AddCommand(repoAddCmd)
	repoAddCmd.Flags().StringP("name", "n", "", "Name of the repository")
	repoAddCmd.Flags().StringP("url", "u", "", "URL of the repository")

	reposCmd.AddCommand(repoRemoveCmd)
	repoRemoveCmd.Flags().StringP("name", "n", "", "Name of the repository to remove")

	reposCmd.AddCommand(repoListCmd)
	reposCmd.AddCommand(repoSyncCmd)
	reposCmd.AddCommand(repoSyncAllCmd)
}
