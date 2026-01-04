package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pr command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
}
