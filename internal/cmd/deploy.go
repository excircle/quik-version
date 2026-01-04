package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/excircle/quik-version/internal/config"
	"github.com/excircle/quik-version/internal/db"
	"github.com/excircle/quik-version/internal/github"
)

var targetBranch string

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the planned version",
	Long: `Deploy executes the version plan from plan.yaml.

This command will:
- Read plan.yaml (fail if missing)
- Authenticate to GitHub
- Create git tag with next_version on latest main commit
- Push tag to GitHub
- If build_management is enabled, trigger buildah container build
- Update qv.db with new version record
- Delete plan.yaml after successful deploy`,
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

		fmt.Printf("Deploying v%s to %s/%s...\n", plan.NextVersion, owner, repo)
		fmt.Println("---")

		// Create GitHub client
		client, err := github.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create GitHub client: %w", err)
		}

		// Get latest commit SHA on target branch
		fmt.Printf("Getting latest commit on '%s'...\n", targetBranch)
		commitSHA, err := client.GetLatestCommitSHA(ctx, owner, repo, targetBranch)
		if err != nil {
			return fmt.Errorf("failed to get latest commit: %w", err)
		}
		fmt.Printf("Commit: %s\n", commitSHA[:7])

		// Create tag
		tagName := "v" + plan.NextVersion
		tagMessage := fmt.Sprintf("Release %s", tagName)

		fmt.Printf("Creating tag '%s'...\n", tagName)
		if err := client.CreateTag(ctx, owner, repo, tagName, commitSHA, tagMessage); err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}

		// Check if build_management is enabled
		if config.GetBuildManagement() {
			fmt.Println("Build management is enabled (buildah integration not yet implemented)")
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Insert new version record
		incrementType := plan.IncrementType
		newVersion := &db.Version{
			Version:       plan.NextVersion,
			TagName:       tagName,
			GitSHA:        commitSHA,
			GitURL:        gitURL,
			IncrementType: &incrementType,
		}

		if err := database.InsertVersion(newVersion); err != nil {
			return fmt.Errorf("failed to record version in database: %w", err)
		}

		// Delete plan.yaml
		if err := os.Remove(planFileName); err != nil {
			fmt.Printf("Warning: failed to delete %s: %v\n", planFileName, err)
		}

		fmt.Println()
		fmt.Println("Deploy successful!")
		fmt.Println("---")
		fmt.Printf("Version: v%s\n", plan.NextVersion)
		fmt.Printf("Tag: %s\n", tagName)
		fmt.Printf("Commit: %s\n", commitSHA)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVar(&targetBranch, "branch", "main", "branch to tag")
}
