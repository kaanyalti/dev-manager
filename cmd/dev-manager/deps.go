package main

import (
	"fmt"
	"os"
	"path/filepath"

	"dev-manager/pkg/config"
	"dev-manager/pkg/deps"

	"github.com/spf13/cobra"
)

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Manage development dependencies",
	Long:  `Manage development dependencies for your workspace.`,
}

var depsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new dependency to the configuration",
	Long: `Add a new dependency to the configuration.
The dependency can be specified with name, version, and source using flags.
Example: dev-manager deps add --name go --version 1.21.0 --source https://go.dev/dl/go1.21.0.darwin-amd64.tar.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("file")
		cfgMgr, err := config.NewManager(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to create config manager: %w", err)
		}

		if err := cfgMgr.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg := cfgMgr.GetConfig()

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		source, _ := cmd.Flags().GetString("source")

		// Validate required flags
		if name == "" {
			return fmt.Errorf("dependency name is required")
		}

		// Check if dependency already exists
		for _, dep := range cfg.Dependencies {
			if dep.Name == name {
				return fmt.Errorf("dependency %s already exists in configuration", name)
			}
		}

		// Create new dependency
		newDep := config.Dependency{
			Name:    name,
			Version: version,
			Source:  source,
		}

		// Add to configuration
		cfg.Dependencies = append(cfg.Dependencies, newDep)

		// Save configuration
		if err := cfgMgr.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Added dependency %s to configuration\n", name)

		// Ask user if they want to install now
		fmt.Print("Would you like to install this dependency now? (Y/n): ")
		var resp string
		fmt.Scanln(&resp)
		if resp == "" || resp == "Y" || resp == "y" {
			depMgr := deps.New(filepath.Join(cfg.WorkspacePath, "deps"))
			if err := depMgr.Install(newDep, false); err != nil {
				return fmt.Errorf("failed to install %s: %w", name, err)
			}
			fmt.Printf("Installed %s\n", name)
		} else {
			fmt.Printf("Dependency %s will be installed during the next sync\n", name)
		}

		return nil
	},
}

var depsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all dependencies",
	Long:  `List all dependencies in the configuration and their installation status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("file")
		cfgMgr, err := config.NewManager(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to create config manager: %w", err)
		}

		if err := cfgMgr.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg := cfgMgr.GetConfig()

		// List all dependencies
		for _, dep := range cfg.Dependencies {
			depPath := filepath.Join(cfg.WorkspacePath, "deps", dep.Name)
			installed := "not installed"
			if _, err := os.Stat(depPath); err == nil {
				installed = "installed"
			}
			fmt.Printf("%s (%s): %s\n", dep.Name, dep.Version, installed)
		}

		return nil
	},
}

var depsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a dependency",
	Long:  `Remove a dependency from the configuration and uninstall it. If no dependency is specified with --name, you will be prompted to select one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("file")
		cfgMgr, err := config.NewManager(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to create config manager: %w", err)
		}

		if err := cfgMgr.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg := cfgMgr.GetConfig()

		// If no dependencies in config, return early
		if len(cfg.Dependencies) == 0 {
			return fmt.Errorf("no dependencies found in configuration")
		}

		var name string
		var index int
		var found bool

		// Check for name flag
		if nameFlag, _ := cmd.Flags().GetString("name"); nameFlag != "" {
			name = nameFlag
			// Find dependency in configuration
			for i, dep := range cfg.Dependencies {
				if dep.Name == name {
					index = i
					found = true
					break
				}
			}
		} else {
			// Interactive prompt for dependency selection
			fmt.Println("Available dependencies:")
			for i, dep := range cfg.Dependencies {
				fmt.Printf("%d. %s (%s)\n", i+1, dep.Name, dep.Version)
			}

			fmt.Print("\nSelect a dependency to remove (number): ")
			var selection int
			fmt.Scanln(&selection)

			if selection < 1 || selection > len(cfg.Dependencies) {
				return fmt.Errorf("invalid selection")
			}

			index = selection - 1
			name = cfg.Dependencies[index].Name
			found = true
		}

		if !found {
			return fmt.Errorf("dependency %s not found in configuration", name)
		}

		// Remove from configuration
		depToRemove := cfg.Dependencies[index]
		cfg.Dependencies = append(cfg.Dependencies[:index], cfg.Dependencies[index+1:]...)

		// Save configuration
		if err := cfgMgr.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		// Uninstall dependency
		depMgr := deps.New(filepath.Join(cfg.WorkspacePath, "deps"))
		if err := depMgr.Remove(depToRemove); err != nil {
			return fmt.Errorf("failed to remove %s: %w", name, err)
		}

		fmt.Printf("Removed dependency %s\n", name)
		return nil
	},
}

var depsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Install all uninstalled dependencies",
	Long:  `Install all dependencies that are in the configuration but not yet installed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("file")
		cfgMgr, err := config.NewManager(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to create config manager: %w", err)
		}

		if err := cfgMgr.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg := cfgMgr.GetConfig()

		// Create dependency manager
		depMgr := deps.New(filepath.Join(cfg.WorkspacePath, "deps"))

		// Install all dependencies
		for _, dep := range cfg.Dependencies {
			if err := depMgr.Install(dep, false); err != nil {
				return fmt.Errorf("failed to install %s: %w", dep.Name, err)
			}
			fmt.Printf("Installed %s\n", dep.Name)
		}

		return nil
	},
}

func init() {
	depsCmd.AddCommand(depsAddCmd)
	depsCmd.AddCommand(depsListCmd)
	depsCmd.AddCommand(depsRemoveCmd)
	depsCmd.AddCommand(depsSyncCmd)

	// Add flags for deps add command
	depsAddCmd.Flags().StringP("name", "n", "", "Name of the dependency")
	depsAddCmd.Flags().StringP("version", "v", "", "Version of the dependency")
	depsAddCmd.Flags().StringP("source", "s", "", "Source URL for the dependency")
	depsAddCmd.MarkFlagRequired("name")

	// Add name flag to depsRemoveCmd
	depsRemoveCmd.Flags().StringP("name", "n", "", "Name of the dependency to remove")

	rootCmd.AddCommand(depsCmd)
}
