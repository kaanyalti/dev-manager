package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"dev-manager/pkg/config"
	"dev-manager/pkg/deps"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Commands for managing dev-manager configuration.`,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration",
	Long: `Validate the current configuration for required fields and structure.
Shows a detailed report of any validation errors found.

Example:
  dev-manager config validate --file config.yaml
  dev-manager config validate -f config.yaml`,
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

		fmt.Printf("Validating configuration at %s...\n\n", mgr.Path())

		if err := cfg.Validate(); err != nil {
			if validationErr, ok := err.(*config.ValidationError); ok {
				fmt.Println(validationErr.Error())
				os.Exit(1)
			}
			log.Fatalf("validation failed: %v", err)
		}

		fmt.Println("Configuration is valid!")
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration",
	Long: `Show the current configuration in a readable format.
Shows workspace path and all managed repositories with their details.

Example:
  dev-manager config show
  dev-manager config show --raw`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("file")
		raw, _ := cmd.Flags().GetBool("raw")

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		if err := mgr.Load(); err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		cfg := mgr.GetConfig()

		if raw {
			// Print raw YAML content
			data, err := yaml.Marshal(cfg)
			if err != nil {
				log.Fatalf("failed to marshal config: %v", err)
			}
			fmt.Println(string(data))
			return
		}

		fmt.Printf("Configuration file: %s\n\n", mgr.Path())
		fmt.Printf("Workspace path: %s\n\n", cfg.WorkspacePath)

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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dev-manager configuration",
	Long: `Initialize dev-manager configuration and install default dependencies.
This will:
1. Create the configuration file
2. Set up the workspace directory
3. Install default dependencies

Example:
  dev-manager init
  dev-manager init --workspace ~/dev`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("file")
		workspace, _ := cmd.Flags().GetString("workspace")
		installDeps, _ := cmd.Flags().GetBool("install-deps")

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

		// Add default dependencies if none exist
		if len(cfg.Dependencies) == 0 {
			cfg.Dependencies = []config.Dependency{
				{
					Name:    "go",
					Version: "1.21.0",
					Source:  "https://go.dev/dl/go1.21.0.darwin-amd64.tar.gz",
				},
				{
					Name:    "node",
					Version: "20.11.1",
					Source:  "https://nodejs.org/dist/v20.11.1/node-v20.11.1-darwin-x64.tar.gz",
				},
			}
		}

		// Save configuration
		if err := mgr.Save(); err != nil {
			log.Fatalf("failed to save configuration: %v", err)
		}

		fmt.Printf("Configuration initialized at %s\n", mgr.Path())
		fmt.Printf("Workspace directory: %s\n", cfg.WorkspacePath)

		// Install dependencies if requested
		if installDeps {
			fmt.Println("\nInstalling dependencies...")
			depMgr := deps.New(filepath.Join(cfg.WorkspacePath, "deps"))
			for _, dep := range cfg.Dependencies {
				if err := depMgr.Install(dep, false); err != nil {
					log.Printf("failed to install %s: %v", dep.Name, err)
					continue
				}
				fmt.Printf("Installed %s\n", dep.Name)
			}
		}
	},
}

func init() {
	// Add config commands
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configShowCmd.Flags().Bool("raw", false, "Show raw YAML content")
	configCmd.AddCommand(configValidateCmd)
	configCmd.PersistentFlags().StringP("file", "f", "", "Path to the configuration file")

	// Add init command
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("workspace", "w", "", "Path to the workspace directory")
	initCmd.Flags().BoolP("install-deps", "i", false, "Install default dependencies")
}
