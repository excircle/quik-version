package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("vet command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(vetCmd)
}
