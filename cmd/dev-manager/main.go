package main

import (
	"fmt"
	"os"

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

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Manage tool configurations",
	Long:  `Commands for managing tool configurations (nvim, tmux, zsh).`,
}

var nvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Manage nvim configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Managing nvim configuration...")
	},
}

var tmuxCmd = &cobra.Command{
	Use:   "tmux",
	Short: "Manage tmux configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Managing tmux configuration...")
	},
}

var zshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Manage zsh configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Managing zsh configuration...")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("file", "f", "", "Path to the configuration file")

	// Add tools commands
	rootCmd.AddCommand(toolsCmd)
	toolsCmd.AddCommand(nvimCmd)
	toolsCmd.AddCommand(tmuxCmd)
	toolsCmd.AddCommand(zshCmd)

	// Add git operations commands
	rootCmd.AddCommand(gitOpsCmd)
}

func main() {
	Execute()
}
