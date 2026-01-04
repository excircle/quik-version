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

var (
	majorFlag bool
	patchFlag bool
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate a version bump plan",
	Long: `Plan reads the current version from qv.db and calculates
the next version based on semantic versioning rules.

By default, increments the MINOR version.
Use --major to increment MAJOR (reset minor/patch).
Use --patch to increment PATCH only.

Generates plan.yaml with the version bump details.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate flags
		if majorFlag && patchFlag {
			return fmt.Errorf("cannot use both --major and --patch flags")
		}

		// Check if database exists
		if !db.Exists() {
			return fmt.Errorf("database not found. Run 'qv init' first")
		}

		// Get git URL from config
		gitURL := config.GetGitURL()
		if gitURL == "" {
			return fmt.Errorf("git_url not configured. Run 'qv init' first")
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Get latest version
		latestVersion, err := database.GetLatestVersion(gitURL)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}

		// Determine increment type and calculate next version
		var incrementType string
		var currentVersion string
		var nextVersion string

		if latestVersion == nil {
			currentVersion = "0.0.0"
			if majorFlag {
				incrementType = "major"
				nextVersion = "1.0.0"
			} else if patchFlag {
				incrementType = "patch"
				nextVersion = "0.0.1"
			} else {
				incrementType = "minor"
				nextVersion = "0.1.0"
			}
		} else {
			currentVersion = latestVersion.Version
			if majorFlag {
				incrementType = "major"
				nextVersion = version.IncrementMajor(currentVersion)
			} else if patchFlag {
				incrementType = "patch"
				nextVersion = version.IncrementPatch(currentVersion)
			} else {
				incrementType = "minor"
				nextVersion = version.IncrementMinor(currentVersion)
			}
		}

		// Create plan
		plan := PlanFile{
			GitURL:         gitURL,
			CurrentVersion: currentVersion,
			NextVersion:    nextVersion,
			IncrementType:  incrementType,
		}

		// Write plan.yaml
		planData, err := yaml.Marshal(&plan)
		if err != nil {
			return fmt.Errorf("failed to marshal plan: %w", err)
		}

		if err := os.WriteFile(planFileName, planData, 0644); err != nil {
			return fmt.Errorf("failed to write plan file: %w", err)
		}

		// Display summary
		fmt.Println("Plan created:")
		fmt.Println("---")
		fmt.Printf("Repository: %s\n", gitURL)
		fmt.Printf("Current Version: v%s\n", currentVersion)
		fmt.Printf("Next Version: v%s\n", nextVersion)
		fmt.Printf("Increment Type: %s\n", incrementType)
		fmt.Println()
		fmt.Printf("Plan saved to %s\n", planFileName)
		fmt.Println("Run 'qv deploy' to apply this plan.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().BoolVar(&majorFlag, "major", false, "increment major version")
	planCmd.Flags().BoolVar(&patchFlag, "patch", false, "increment patch version")
}
