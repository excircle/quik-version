package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/excircle/quik-version/internal/config"
	"github.com/excircle/quik-version/internal/db"
	"github.com/excircle/quik-version/internal/github"
)

var vetCmd = &cobra.Command{
	Use:   "vet",
	Short: "Validate git tags against local database",
	Long: `Vet checks quik.conf for git_url and validates that
local qv.db matches the remote GitHub tags.

This command will:
- Load config and validate git_url exists
- Fetch latest tags from GitHub
- Compare with local qv.db
- Report discrepancies and offer reconciliation options`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Check if database exists
		if !db.Exists() {
			return fmt.Errorf("database not found. Run 'qv init' first")
		}

		// Get git URL from config
		gitURL := config.GetGitURL()
		if gitURL == "" {
			return fmt.Errorf("git_url not configured. Run 'qv init' first")
		}

		// Parse owner and repo from URL
		owner, repo, err := github.ParseRepoURL(gitURL)
		if err != nil {
			return fmt.Errorf("failed to parse git URL: %w", err)
		}

		fmt.Printf("Repository: %s/%s\n", owner, repo)
		fmt.Println("---")

		// Create GitHub client
		client, err := github.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create GitHub client: %w", err)
		}

		// Fetch remote tags
		fmt.Println("Fetching remote tags...")
		remoteTags, err := client.ListTags(ctx, owner, repo)
		if err != nil {
			return fmt.Errorf("failed to fetch remote tags: %w", err)
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Get local versions
		localVersions, err := database.GetAllVersions(gitURL)
		if err != nil {
			return fmt.Errorf("failed to get local versions: %w", err)
		}

		// Build maps for comparison
		remoteTagMap := make(map[string]string) // tag name -> SHA
		for _, tag := range remoteTags {
			remoteTagMap[tag.Name] = tag.SHA
		}

		localTagMap := make(map[string]string) // tag name -> SHA
		for _, v := range localVersions {
			localTagMap[v.TagName] = v.GitSHA
		}

		// Find discrepancies
		var remoteOnly []string
		var localOnly []string
		var mismatched []string

		for tagName, remoteSHA := range remoteTagMap {
			if localSHA, exists := localTagMap[tagName]; !exists {
				remoteOnly = append(remoteOnly, tagName)
			} else if localSHA != remoteSHA {
				mismatched = append(mismatched, tagName)
			}
		}

		for tagName := range localTagMap {
			if _, exists := remoteTagMap[tagName]; !exists {
				localOnly = append(localOnly, tagName)
			}
		}

		// Report findings
		fmt.Printf("\nRemote tags: %d\n", len(remoteTags))
		fmt.Printf("Local versions: %d\n", len(localVersions))
		fmt.Println()

		if len(remoteOnly) == 0 && len(localOnly) == 0 && len(mismatched) == 0 {
			fmt.Println("✓ Database is in sync with remote.")
			return nil
		}

		if len(remoteOnly) > 0 {
			fmt.Printf("Tags on remote but not in local DB (%d):\n", len(remoteOnly))
			for _, tag := range remoteOnly {
				fmt.Printf("  + %s\n", tag)
			}
			fmt.Println()
		}

		if len(localOnly) > 0 {
			fmt.Printf("Tags in local DB but not on remote (%d):\n", len(localOnly))
			for _, tag := range localOnly {
				fmt.Printf("  - %s\n", tag)
			}
			fmt.Println()
		}

		if len(mismatched) > 0 {
			fmt.Printf("Tags with mismatched SHAs (%d):\n", len(mismatched))
			for _, tag := range mismatched {
				fmt.Printf("  ! %s\n", tag)
			}
			fmt.Println()
		}

		// Offer reconciliation
		if len(remoteOnly) > 0 {
			fmt.Print("Sync remote tags to local DB? (y/n): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				for _, tagName := range remoteOnly {
					sha := remoteTagMap[tagName]
					version := strings.TrimPrefix(tagName, "v")
					err := database.InsertVersion(&db.Version{
						Version: version,
						TagName: tagName,
						GitSHA:  sha,
						GitURL:  gitURL,
					})
					if err != nil {
						fmt.Printf("  Failed to add %s: %v\n", tagName, err)
					} else {
						fmt.Printf("  Added %s\n", tagName)
					}
				}
				fmt.Println("Sync complete.")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(vetCmd)
}
