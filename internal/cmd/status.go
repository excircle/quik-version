package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/excircle/quik-version/internal/config"
	"github.com/excircle/quik-version/internal/db"
	"github.com/excircle/quik-version/internal/version"
)

const planFileName = "plan.yaml"

// PlanFile represents the structure of plan.yaml
type PlanFile struct {
	GitURL         string `yaml:"git_url"`
	CurrentVersion string `yaml:"current_version"`
	NextVersion    string `yaml:"next_version"`
	IncrementType  string `yaml:"increment_type"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display current version status",
	Long: `Status reports the latest versioning data from qv.db.

This command will:
- Query qv.db for latest version record
- Display current version, last tag, commit SHA, timestamp
- Show pending changes if plan.yaml exists
- Show what would be created if a plan were run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if database exists
		if !db.Exists() {
			return fmt.Errorf("database not found. Run 'qv init' first")
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Get git URL from config
		gitURL := config.GetGitURL()
		if gitURL == "" {
			return fmt.Errorf("git_url not configured. Run 'qv init' first")
		}

		fmt.Printf("Repository: %s\n", gitURL)
		fmt.Println("---")

		// Get latest version
		latestVersion, err := database.GetLatestVersion(gitURL)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}

		if latestVersion == nil {
			fmt.Println("No versions recorded yet.")
			fmt.Println()
			fmt.Println("Next version (if plan is run):")
			fmt.Println("  Minor: v0.1.0")
			fmt.Println("  Major: v1.0.0")
			fmt.Println("  Patch: v0.0.1")
		} else {
			fmt.Printf("Current Version: %s\n", latestVersion.Version)
			fmt.Printf("Tag: %s\n", latestVersion.TagName)
			fmt.Printf("Commit SHA: %s\n", latestVersion.GitSHA)
			fmt.Printf("Created: %s\n", latestVersion.CreatedAt)

			// Show what next versions would be
			nextMajor := version.IncrementMajor(latestVersion.Version)
			nextMinor := version.IncrementMinor(latestVersion.Version)
			nextPatch := version.IncrementPatch(latestVersion.Version)

			fmt.Println()
			fmt.Println("Next version (if plan is run):")
			fmt.Printf("  Minor (default): v%s\n", nextMinor)
			fmt.Printf("  Major (--major): v%s\n", nextMajor)
			fmt.Printf("  Patch (--patch): v%s\n", nextPatch)
		}

		// Check for pending plan
		if _, err := os.Stat(planFileName); err == nil {
			fmt.Println()
			fmt.Println("---")
			fmt.Println("PENDING PLAN:")

			planData, err := os.ReadFile(planFileName)
			if err != nil {
				return fmt.Errorf("failed to read plan file: %w", err)
			}

			var plan PlanFile
			if err := yaml.Unmarshal(planData, &plan); err != nil {
				return fmt.Errorf("failed to parse plan file: %w", err)
			}

			fmt.Printf("  Current: %s\n", plan.CurrentVersion)
			fmt.Printf("  Next: %s\n", plan.NextVersion)
			fmt.Printf("  Type: %s\n", plan.IncrementType)
			fmt.Println()
			fmt.Println("Run 'qv deploy' to apply this plan.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
