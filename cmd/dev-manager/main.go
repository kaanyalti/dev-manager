package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"dev-manager/internal/ssh"
	"dev-manager/pkg/config"
	"dev-manager/pkg/git"

	"github.com/atotto/clipboard"
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

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "Manage repositories",
	Long:  `Manage repositories in your development environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
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

		cfgPath, _ := cmd.Flags().GetString("config")
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
				log.Fatalf("Failed to clone repository: %v", err)
			}
			fmt.Printf("Successfully cloned repository to %s\n", newRepo.Path)
		} else {
			fmt.Println("Repository will be cloned during the next sync.")
		}
	},
}

var repoRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a repository from management",
	Long: `Remove a repository from dev-manager.
This will remove the repository from the configuration and optionally delete the repository directory.

Example:
  dev-manager repos remove --name my-project
  dev-manager repos remove (interactive selection)`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("config")
		repoName, _ := cmd.Flags().GetString("name")

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		if err := mgr.Load(); err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		cfg := mgr.GetConfig()

		var selectedRepo *config.Repository
		var repoIndex int
		if repoName == "" {
			// List available repositories and prompt for selection
			if len(cfg.Repositories) == 0 {
				fmt.Println("No repositories found in configuration.")
				os.Exit(1)
			}
			fmt.Println("Available repositories:")
			for i, repo := range cfg.Repositories {
				fmt.Printf("  [%d] %s (%s)\n", i+1, repo.Name, repo.URL)
			}
			fmt.Print("Select a repository to remove (number): ")
			var idx int
			_, err = fmt.Scanln(&idx)
			if err != nil || idx < 1 || idx > len(cfg.Repositories) {
				fmt.Println("Invalid selection.")
				os.Exit(1)
			}
			repoIndex = idx - 1
			selectedRepo = &cfg.Repositories[repoIndex]
		} else {
			// Find the repository by name
			for i, repo := range cfg.Repositories {
				if repo.Name == repoName {
					repoIndex = i
					selectedRepo = &cfg.Repositories[i]
					break
				}
			}
			if selectedRepo == nil {
				log.Fatalf("repository with name '%s' is not managed by dev-manager", repoName)
			}
		}

		// Check if directory exists
		if _, err := os.Stat(selectedRepo.Path); err == nil {
			fmt.Printf("Repository directory exists at: %s\n", selectedRepo.Path)
			fmt.Print("Would you like to delete the repository directory? (y/N): ")
			var resp string
			fmt.Scanln(&resp)
			if resp == "y" || resp == "Y" {
				if err := os.RemoveAll(selectedRepo.Path); err != nil {
					log.Fatalf("failed to remove repository directory: %v", err)
				}
				fmt.Printf("Removed repository directory: %s\n", selectedRepo.Path)
			} else {
				fmt.Println("Keeping repository directory.")
			}
		}

		// Remove from config using the stored index
		cfg.Repositories = append(cfg.Repositories[:repoIndex], cfg.Repositories[repoIndex+1:]...)

		// Save configuration
		if err := mgr.Save(); err != nil {
			log.Fatalf("failed to save configuration: %v", err)
		}

		fmt.Printf("Removed repository '%s' from configuration\n", selectedRepo.Name)
	},
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed repositories",
	Long: `List all repositories managed by dev-manager.
Shows repository name, URL, path, branch, and last sync time.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("config")

		mgr, err := config.NewManager(cfgPath)
		if err != nil {
			log.Fatalf("failed to create config manager: %v", err)
		}

		if err := mgr.Load(); err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		cfg := mgr.GetConfig()

		if len(cfg.Repositories) == 0 {
			fmt.Println("No repositories found in configuration.")
			return
		}

		fmt.Printf("Found %d repositories:\n\n", len(cfg.Repositories))
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

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Manage SSH keys and agent for your dev environment",
	Long: `Manage SSH keys and the ssh-agent for your development environment. 
Provides commands to generate, list, add, remove, and inspect SSH keys and agent status.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
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
			fmt.Println("No SSH key pairs found.")
			var resp string
			fmt.Print("Would you like to generate a new SSH key pair? (Y/n): ")
			fmt.Scanln(&resp)
			if resp != "" && resp != "Y" && resp != "y" {
				fmt.Println("Skipping SSH key generation.")
				os.Exit(0)
			}

			// Prompt for algorithm
			algo := "ed25519"
			fmt.Print("Enter key algorithm (ed25519/rsa/ecdsa) [ed25519]: ")
			var inputAlgo string
			fmt.Scanln(&inputAlgo)
			if inputAlgo != "" {
				algo = inputAlgo
			}

			// Prompt for name
			fmt.Print("Enter a name for the key (optional): ")
			var keyName string
			fmt.Scanln(&keyName)

			keyPath, err := mgr.GenerateKey(algo, keyName)
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

var sshGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new SSH key pair",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

		algo, _ := cmd.Flags().GetString("algo")
		name, _ := cmd.Flags().GetString("name")

		if algo == "" {
			fmt.Print("Enter key algorithm (ed25519/rsa/ecdsa) [ed25519]: ")
			fmt.Scanln(&algo)
			if algo == "" {
				algo = "ed25519"
			}
		}

		if name == "" {
			fmt.Print("Enter a name for the key (optional): ")
			fmt.Scanln(&name)
		}

		// Determine intended key path
		sshDir := filepath.Join(mgr.HomeDir, ".ssh")
		keyFile := "id_" + algo
		if name != "" {
			keyFile = name + "_id_" + algo
		}
		keyPath := filepath.Join(sshDir, keyFile)
		if _, err := os.Stat(keyPath); err == nil {
			fmt.Printf("A key with this name and algorithm already exists: %s\n", keyPath)
			fmt.Println("Aborting to avoid overwriting existing key.")
			os.Exit(1)
		}

		keyPath, err = mgr.GenerateKey(algo, name)
		if err != nil {
			log.Fatalf("Failed to generate SSH key: %v", err)
		}
		fmt.Printf("New SSH key pair generated: %s\n", keyPath)

		if err := mgr.PrintPublicKey(keyPath); err != nil {
			log.Printf("Could not print public key: %v", err)
		}

		// Offer to add to agent
		var resp string
		fmt.Print("Would you like to add this key to the ssh-agent? (Y/n): ")
		fmt.Scanln(&resp)
		if resp == "" || resp == "Y" || resp == "y" {
			if err := mgr.AddKeyToAgent(keyPath); err != nil {
				log.Fatalf("Failed to add key to agent: %v", err)
			}
			fmt.Println("Key added to agent.")
		} else {
			fmt.Println("Key not added to agent.")
		}
	},
}

var sshListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available SSH key pairs and agent-loaded keys",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

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

var sshAddAgentCmd = &cobra.Command{
	Use:   "add-agent",
	Short: "Add a private SSH key to the ssh-agent",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

		keyPath, _ := cmd.Flags().GetString("key")
		if keyPath == "" {
			// List available keys and prompt for selection
			keys, err := mgr.ListPrivateKeys()
			if err != nil {
				log.Fatalf("Failed to list SSH keys: %v", err)
			}
			if len(keys) == 0 {
				fmt.Println("No SSH keys found in ~/.ssh.")
				os.Exit(1)
			}
			fmt.Println("Available SSH keys:")
			for i, k := range keys {
				fmt.Printf("  [%d] %s\n", i+1, k)
			}
			fmt.Print("Select a key to add to the agent (number): ")
			var idx int
			_, err = fmt.Scanln(&idx)
			if err != nil || idx < 1 || idx > len(keys) {
				fmt.Println("Invalid selection.")
				os.Exit(1)
			}
			keyPath = keys[idx-1]
		}

		if err := mgr.AddKeyToAgent(keyPath); err != nil {
			log.Fatalf("Failed to add key to agent: %v", err)
		}
		fmt.Printf("Key added to agent: %s\n", keyPath)
	},
}

var sshStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show SSH environment status (tooling, agent, keys, agent keys)",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

		fmt.Println("Checking required SSH tools...")
		if err := mgr.CheckTools(); err != nil {
			fmt.Printf("  [!] %v\n", err)
		} else {
			fmt.Println("  [✓] All required SSH tools are installed.")
		}

		fmt.Print("Checking if ssh-agent is running... ")
		if mgr.IsAgentRunning() {
			fmt.Println("[✓] ssh-agent is running.")
		} else {
			fmt.Println("[!] ssh-agent is NOT running.")
		}

		fmt.Println("\nPrivate SSH keys in ~/.ssh:")
		keys, err := mgr.ListPrivateKeys()
		if err != nil {
			fmt.Printf("  [!] Failed to list SSH keys: %v\n", err)
		} else if len(keys) == 0 {
			fmt.Println("  (none found)")
		} else {
			for _, k := range keys {
				fmt.Println("  ", k)
			}
		}

		fmt.Println("\nKeys loaded in ssh-agent:")
		agentKeys, err := mgr.ListAgentKeys()
		if err != nil {
			fmt.Printf("  [!] Failed to list agent keys: %v\n", err)
		} else if len(agentKeys) == 0 {
			fmt.Println("  (none loaded)")
		} else {
			for _, k := range agentKeys {
				fmt.Println("  ", k)
			}
		}
	},
}

var sshPrintPublicCmd = &cobra.Command{
	Use:   "print-public",
	Short: "Print the public key for a given private SSH key",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

		keyPath, _ := cmd.Flags().GetString("key")
		if keyPath == "" {
			// List available keys and prompt for selection
			keys, err := mgr.ListPrivateKeys()
			if err != nil {
				log.Fatalf("Failed to list SSH keys: %v", err)
			}
			if len(keys) == 0 {
				fmt.Println("No SSH keys found in ~/.ssh.")
				os.Exit(1)
			}
			fmt.Println("Available SSH keys:")
			for i, k := range keys {
				fmt.Printf("  [%d] %s\n", i+1, k)
			}
			fmt.Print("Select a key to print its public key (number): ")
			var idx int
			_, err = fmt.Scanln(&idx)
			if err != nil || idx < 1 || idx > len(keys) {
				fmt.Println("Invalid selection.")
				os.Exit(1)
			}
			keyPath = keys[idx-1]
		}

		if err := mgr.PrintPublicKey(keyPath); err != nil {
			log.Fatalf("Could not print public key: %v", err)
		}
	},
}

var sshCopyPublicCmd = &cobra.Command{
	Use:   "copy-public",
	Short: "Copy the public key for a given private SSH key to the clipboard",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

		keyPath, _ := cmd.Flags().GetString("key")
		if keyPath == "" {
			// List available keys and prompt for selection
			keys, err := mgr.ListPrivateKeys()
			if err != nil {
				log.Fatalf("Failed to list SSH keys: %v", err)
			}
			if len(keys) == 0 {
				fmt.Println("No SSH keys found in ~/.ssh.")
				os.Exit(1)
			}
			fmt.Println("Available SSH keys:")
			for i, k := range keys {
				fmt.Printf("  [%d] %s\n", i+1, k)
			}
			fmt.Print("Select a key to copy its public key (number): ")
			var idx int
			_, err = fmt.Scanln(&idx)
			if err != nil || idx < 1 || idx > len(keys) {
				fmt.Println("Invalid selection.")
				os.Exit(1)
			}
			keyPath = keys[idx-1]
		}

		pubPath := keyPath + ".pub"
		data, err := os.ReadFile(pubPath)
		if err != nil {
			log.Fatalf("Could not read public key: %v", err)
		}
		if err := clipboard.WriteAll(string(data)); err != nil {
			log.Fatalf("Failed to copy public key to clipboard: %v", err)
		}
		fmt.Printf("Public key copied to clipboard: %s\n", pubPath)
	},
}

var sshRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a private SSH key from the agent and delete it from disk",
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := ssh.NewSSHManager()
		if err != nil {
			log.Fatalf("Failed to initialize SSH manager: %v", err)
		}

		keyPath, _ := cmd.Flags().GetString("key")
		if keyPath == "" {
			// List available keys and prompt for selection
			keys, err := mgr.ListPrivateKeys()
			if err != nil {
				log.Fatalf("Failed to list SSH keys: %v", err)
			}
			if len(keys) == 0 {
				fmt.Println("No SSH keys found in ~/.ssh.")
				os.Exit(1)
			}
			fmt.Println("Available SSH keys:")
			for i, k := range keys {
				fmt.Printf("  [%d] %s\n", i+1, k)
			}
			fmt.Print("Select a key to remove (number): ")
			var idx int
			_, err = fmt.Scanln(&idx)
			if err != nil || idx < 1 || idx > len(keys) {
				fmt.Println("Invalid selection.")
				os.Exit(1)
			}
			keyPath = keys[idx-1]
		}

		fmt.Printf("Are you sure you want to remove the key %s from the agent and delete it from disk? (y/N): ", keyPath)
		var resp string
		fmt.Scanln(&resp)
		if resp != "y" && resp != "Y" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}

		// Remove from agent
		cmdRm := exec.Command("ssh-add", "-d", keyPath)
		cmdRm.Stdin = os.Stdin
		cmdRm.Stdout = os.Stdout
		cmdRm.Stderr = os.Stderr
		_ = cmdRm.Run() // ignore error if not loaded

		// Delete private and public key files
		pubPath := keyPath + ".pub"
		if err := os.Remove(keyPath); err != nil {
			fmt.Printf("Failed to delete private key: %v\n", err)
		} else {
			fmt.Printf("Deleted private key: %s\n", keyPath)
		}
		if err := os.Remove(pubPath); err != nil {
			fmt.Printf("Failed to delete public key: %v\n", err)
		} else {
			fmt.Printf("Deleted public key: %s\n", pubPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(reposCmd)

	// Flags for init command
	initCmd.Flags().StringP("config", "c", "", "Path to configuration file")
	initCmd.Flags().StringP("workspace", "w", "", "Workspace directory for cloning repositories")

	reposCmd.AddCommand(repoAddCmd)
	reposCmd.AddCommand(repoRemoveCmd)
	reposCmd.AddCommand(repoListCmd)

	// Flags for add-repo command
	repoAddCmd.Flags().StringP("name", "n", "", "Name of the repository")
	repoAddCmd.Flags().StringP("url", "u", "", "Repository URL to clone")

	// Flags for remove-repo command
	repoRemoveCmd.Flags().StringP("name", "n", "", "Name of the repository to remove")

	// SSH command group and subcommands
	sshCmd.AddCommand(sshInitCmd)
	sshCmd.AddCommand(sshGenerateCmd)
	sshCmd.AddCommand(sshListCmd)
	sshCmd.AddCommand(sshAddAgentCmd)
	sshCmd.AddCommand(sshStatusCmd)
	sshCmd.AddCommand(sshPrintPublicCmd)
	sshCmd.AddCommand(sshCopyPublicCmd)
	sshCmd.AddCommand(sshRemoveCmd)
	rootCmd.AddCommand(sshCmd)

	sshGenerateCmd.Flags().String("algo", "", "Key algorithm (ed25519, rsa, ecdsa)")
	sshGenerateCmd.Flags().String("name", "", "Name for the key (optional)")
	sshAddAgentCmd.Flags().String("key", "", "Path to the private key to add to the agent")
	sshPrintPublicCmd.Flags().String("key", "", "Path to the private key to print its public key")
	sshCopyPublicCmd.Flags().String("key", "", "Path to the private key to copy its public key")
	sshRemoveCmd.Flags().String("key", "", "Path to the private key to remove and delete")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
