package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display current version status",
	Long: `Status reports the latest versioning data from qv.db.

This command will:
- Query qv.db for latest version record
- Display current version, last tag, commit SHA, timestamp
- Show pending changes if plan.yaml exists
- Show what would be created if a plan were run`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("status command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
