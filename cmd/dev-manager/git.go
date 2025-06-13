package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var gitOpsCmd = &cobra.Command{
	Use:   "git-ops",
	Short: "Git operations and workflow",
	Long:  `Commands for performing git operations like commit, push, and other git workflow actions.`,
}

var gitCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Stage, commit, and push changes with an LLM-generated commit message",
	Long: `Stage, commit, and push changes with an LLM-generated commit message.
If no custom message is provided, an LLM will generate one based on the changes.
You can review the changes before committing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		customMsg, _ := cmd.Flags().GetString("message")
		noPush, _ := cmd.Flags().GetBool("no-push")
		noLLM, _ := cmd.Flags().GetBool("no-llm")

		// Stage all changes
		stageCmd := exec.Command("git", "add", ".")
		stageCmd.Stdout = os.Stdout
		stageCmd.Stderr = os.Stderr
		if err := stageCmd.Run(); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}

		// Get staged changes
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

		// Get commit message
		var commitMsg string
		if customMsg != "" {
			commitMsg = customMsg
		} else if !noLLM {
			// Generate commit message using OpenAI
			apiKey := os.Getenv("OPENAI_API_KEY")
			if apiKey == "" {
				return fmt.Errorf("OPENAI_API_KEY environment variable is required for LLM commit messages")
			}

			commitMsg, err = generateCommitMessageWithLLM(string(diffOutput), apiKey)
			if err != nil {
				return fmt.Errorf("failed to generate commit message: %w", err)
			}

			// Show proposed commit message
			fmt.Println("\nProposed commit message:")
			fmt.Println(commitMsg)
			fmt.Println("\nDo you want to use this commit message? (y/N): ")

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
		} else {
			// Prompt for manual commit message
			fmt.Print("\nEnter commit message: ")
			commitMsg, err = reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read commit message: %w", err)
			}
			commitMsg = strings.TrimSpace(commitMsg)
		}

		// Commit changes
		commitCmd := exec.Command("git", "commit", "-m", commitMsg)
		commitCmd.Stdout = os.Stdout
		commitCmd.Stderr = os.Stderr
		if err := commitCmd.Run(); err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}

		// Push changes if not disabled
		if !noPush {
			pushCmd := exec.Command("git", "push")
			pushCmd.Stdout = os.Stdout
			pushCmd.Stderr = os.Stderr
			if err := pushCmd.Run(); err != nil {
				return fmt.Errorf("failed to push changes: %w", err)
			}
			fmt.Println("Changes committed and pushed successfully!")
		} else {
			fmt.Println("Changes committed successfully!")
		}

		return nil
	},
}

func init() {
	gitOpsCmd.AddCommand(gitCommitCmd)

	// Add flags
	gitCommitCmd.Flags().StringP("message", "m", "", "Custom commit message")
	gitCommitCmd.Flags().Bool("no-push", false, "Don't push after commit")
	gitCommitCmd.Flags().Bool("no-llm", false, "Don't use LLM for commit message")
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
