package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/excircle/quik-version/internal/config"
	"github.com/excircle/quik-version/internal/github"
)

var baseBranch string

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a pull request with version info",
	Long: `PR creates a GitHub pull request from the current branch to main.

This command will:
- Read plan.yaml (fail if missing)
- Detect current branch name
- Authenticate to GitHub
- Create PR with version details in title and body
- Display PR URL`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Check if plan.yaml exists
		if _, err := os.Stat(planFileName); os.IsNotExist(err) {
			return fmt.Errorf("plan.yaml not found. Run 'qv plan' first")
		}

		// Read plan.yaml
		planData, err := os.ReadFile(planFileName)
		if err != nil {
			return fmt.Errorf("failed to read plan file: %w", err)
		}

		var plan PlanFile
		if err := yaml.Unmarshal(planData, &plan); err != nil {
			return fmt.Errorf("failed to parse plan file: %w", err)
		}

		// Get current branch name
		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branchOutput, err := branchCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		currentBranch := strings.TrimSpace(string(branchOutput))

		if currentBranch == baseBranch {
			return fmt.Errorf("cannot create PR from %s to %s", baseBranch, baseBranch)
		}

		// Get git URL from config
		gitURL := config.GetGitURL()
		if gitURL == "" {
			return fmt.Errorf("git_url not configured. Run 'qv init' first")
		}

		// Parse owner and repo
		owner, repo, err := github.ParseRepoURL(gitURL)
		if err != nil {
			return fmt.Errorf("failed to parse git URL: %w", err)
		}

		// Create GitHub client
		client, err := github.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create GitHub client: %w", err)
		}

		// Build PR title and body
		title := fmt.Sprintf("Release v%s", plan.NextVersion)
		body := fmt.Sprintf(`## Version Bump

**Current Version:** v%s
**Next Version:** v%s
**Increment Type:** %s

---
*Created by qv (Quik Version)*
`, plan.CurrentVersion, plan.NextVersion, plan.IncrementType)

		fmt.Printf("Creating PR from '%s' to '%s'...\n", currentBranch, baseBranch)

		// Create PR
		pr, err := client.CreatePR(ctx, owner, repo, title, body, currentBranch, baseBranch)
		if err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}

		fmt.Println()
		fmt.Println("Pull request created:")
		fmt.Println("---")
		fmt.Printf("Title: %s\n", pr.Title)
		fmt.Printf("Number: #%d\n", pr.Number)
		fmt.Printf("URL: %s\n", pr.URL)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.Flags().StringVar(&baseBranch, "base", "main", "base branch for the PR")
}
