//go:build mage

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Build compiles the dev-manager binary
func Build() error {
	fmt.Println("Building dev-manager...")

	// Create bin directory if it doesn't exist
	if err := os.MkdirAll("bin", 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", filepath.Join("bin", "dev-manager"), "./cmd/dev-manager")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	fmt.Println("Build completed successfully!")
	return nil
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	return os.RemoveAll("bin")
}

// GC is an alias for GitCommit that stages, commits, and pushes changes with an LLM-generated commit message
func GC() error {
	return GitCommit()
}

// GitCommit stages, commits, and pushes changes with an LLM-generated commit message
func GitCommit() error {
	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Stage all changes
	cmd := exec.Command("git", "add", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Get staged changes for commit message
	diffCmd := exec.Command("git", "diff", "--cached")
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	// Get list of changed files
	filesCmd := exec.Command("git", "diff", "--cached", "--name-only")
	filesOutput, err := filesCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(filesOutput)), "\n")
	if len(changedFiles) == 0 {
		return fmt.Errorf("no changes to commit")
	}

	// Interactive file review loop
	reader := bufio.NewReader(os.Stdin)
	for {
		// Show changed files
		fmt.Println("\nChanged files:")
		for i, file := range changedFiles {
			fmt.Printf("%d. %s\n", i+1, file)
		}

		// Ask for file number to review
		fmt.Print("\nEnter file number to review (or press enter to continue): ")
		fileNumStr, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read file number: %w", err)
		}

		fileNumStr = strings.TrimSpace(fileNumStr)
		if fileNumStr == "" {
			break
		}

		fileNum, err := strconv.Atoi(fileNumStr)
		if err != nil || fileNum < 1 || fileNum > len(changedFiles) {
			fmt.Println("Invalid file number")
			continue
		}

		// Show diff for selected file
		fileDiffCmd := exec.Command("git", "diff", "--cached", "--", changedFiles[fileNum-1])
		fileDiffOutput, err := fileDiffCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get file diff: %w", err)
		}

		fmt.Printf("\nDiff for %s:\n", changedFiles[fileNum-1])
		fmt.Println(string(fileDiffOutput))
	}

	// Generate commit message using OpenAI
	commitMsg, err := generateCommitMessageWithLLM(string(diffOutput), apiKey)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Show proposed commit message
	fmt.Println("\nProposed commit message:")
	fmt.Println(commitMsg)
	fmt.Println("\nDo you want to commit and push these changes? (y/N): ")

	// Get user confirmation
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Aborted.")
		return nil
	}

	// Commit changes
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push changes
	pushCmd := exec.Command("git", "push")
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	fmt.Println("Changes committed and pushed successfully!")
	return nil
}

// generateCommitMessageWithLLM uses OpenAI to generate a commit message based on the changes
func generateCommitMessageWithLLM(diff, apiKey string) (string, error) {
	client := openai.NewClient(apiKey)

	// Prepare the prompt
	prompt := fmt.Sprintf(`Generate a concise and descriptive commit message for the following changes.
Follow conventional commit format (e.g., feat:, fix:, chore:, etc.).
Focus on the main changes and their impact.
Keep the message under 72 characters.

Changes:
%s`, diff)

	// Create the completion request
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant that generates commit messages. Be concise and follow conventional commit format.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	// Get the completion
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to get completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
