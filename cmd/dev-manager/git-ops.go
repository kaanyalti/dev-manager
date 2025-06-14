package main

import (
	"bufio"
	"context"
	"encoding/json"
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
	Short: "Git workflow enhancements",
	Long:  `Commands for git workflow enhancements like LLM-powered commit messages and other git operation improvements.`,
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

var gitReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Analyze PR comments and provide LLM-powered suggestions",
	Long: `Analyze PR comments and provide LLM-powered suggestions for addressing review feedback.
This command will:
1. Fetch PR comments from the current repository
2. Analyze comments using LLM
3. Provide suggestions for addressing each comment
4. Help generate responses to reviewers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get PR number from flag
		prNumber, _ := cmd.Flags().GetInt("pr")
		if prNumber == 0 {
			// Get current branch name
			branchCmd := exec.Command("git", "branch", "--show-current")
			branchOutput, err := branchCmd.Output()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}
			branchName := strings.TrimSpace(string(branchOutput))

			// Search for PRs associated with current branch and user
			searchCmd := exec.Command("gh", "search", "prs", "--json", "number,title", "--jq", ".[0]", fmt.Sprintf("head:%s", branchName), "is:open")
			searchOutput, err := searchCmd.Output()
			if err == nil && len(searchOutput) > 0 {
				var pr struct {
					Number int    `json:"number"`
					Title  string `json:"title"`
				}
				if err := json.Unmarshal(searchOutput, &pr); err == nil {
					fmt.Printf("Found PR #%d: %s\nUse this PR? [y/N]: ", pr.Number, pr.Title)
					reader := bufio.NewReader(os.Stdin)
					response, err := reader.ReadString('\n')
					if err != nil {
						return fmt.Errorf("failed to read response: %w", err)
					}
					if strings.ToLower(strings.TrimSpace(response)) == "y" {
						prNumber = pr.Number
					}
				}
			}

			// If no PR number yet, prompt user
			if prNumber == 0 {
				fmt.Print("Enter PR number: ")
				reader := bufio.NewReader(os.Stdin)
				prStr, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read PR number: %w", err)
				}
				prNumber, err = strconv.Atoi(strings.TrimSpace(prStr))
				if err != nil {
					return fmt.Errorf("invalid PR number: %w", err)
				}
			}
		}

		// Validate PR exists
		validateCmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", prNumber), "--json", "number")
		if err := validateCmd.Run(); err != nil {
			return fmt.Errorf("PR #%d not found or not accessible: %w", prNumber, err)
		}

		// Get PR details including comments, diff, and metadata
		prCmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", prNumber), "--json", "title,body,comments,reviewComments,commits,files")
		prOutput, err := prCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get PR details: %w", err)
		}

		// Generate suggestions using OpenAI
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}

		suggestions, err := generatePRReviewSuggestions(string(prOutput), apiKey)
		if err != nil {
			return fmt.Errorf("failed to generate suggestions: %w", err)
		}

		// Print suggestions
		fmt.Println("\nPR Review Suggestions:")
		fmt.Println(suggestions)

		return nil
	},
}

func init() {
	gitOpsCmd.AddCommand(gitCommitCmd)
	gitOpsCmd.AddCommand(gitReviewCmd)

	// Add flags
	gitCommitCmd.Flags().StringP("message", "m", "", "Custom commit message")
	gitCommitCmd.Flags().Bool("no-push", false, "Don't push after commit")
	gitCommitCmd.Flags().Bool("no-llm", false, "Don't use LLM for commit message")

	gitReviewCmd.Flags().IntP("pr", "p", 0, "PR number (optional, will try to detect from branch name)")
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

// generatePRReviewSuggestions uses OpenAI to generate suggestions based on PR comments
func generatePRReviewSuggestions(prData, apiKey string) (string, error) {
	client := openai.NewClient(apiKey)

	// Parse PR data
	var pr struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		Comments []struct {
			Body string `json:"body"`
		} `json:"comments"`
		ReviewComments []struct {
			Body string `json:"body"`
		} `json:"reviewComments"`
		Files []struct {
			Path      string `json:"path"`
			Additions int    `json:"additions"`
			Deletions int    `json:"deletions"`
			Changes   int    `json:"changes"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(prData), &pr); err != nil {
		return "", fmt.Errorf("failed to parse PR data: %w", err)
	}

	// Prepare the prompt
	prompt := fmt.Sprintf(`Analyze these PR comments and provide suggestions for addressing them.
For each comment:
1. Summarize the main point
2. Suggest specific code changes if applicable
3. Provide a draft response to the reviewer
4. Categorize the comment (e.g., bug, enhancement, style, etc.)

PR Title: %s
PR Description: %s

PR Comments:
%s

PR Review Comments:
%s

Changed Files:
%s`,
		pr.Title,
		pr.Body,
		formatComments(pr.Comments),
		formatComments(pr.ReviewComments),
		formatFiles(pr.Files))

	// Create the completion request
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant that analyzes PR comments and provides actionable suggestions. Be specific and practical in your recommendations.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   1000,
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

// formatComments formats a list of comments into a readable string
func formatComments(comments []struct {
	Body string `json:"body"`
}) string {
	var result strings.Builder
	for i, comment := range comments {
		result.WriteString(fmt.Sprintf("Comment %d:\n%s\n\n", i+1, comment.Body))
	}
	return result.String()
}

// formatFiles formats a list of changed files into a readable string
func formatFiles(files []struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
}) string {
	var result strings.Builder
	for _, file := range files {
		result.WriteString(fmt.Sprintf("%s: +%d -%d (%d changes)\n",
			file.Path, file.Additions, file.Deletions, file.Changes))
	}
	return result.String()
}
