package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Quik Version configuration",
	Long: `Initialize Quik Version by creating quik.conf and qv.db.

This command will:
- Check for existing quik.conf and prompt to overwrite or skip
- Prompt for git_url
- Prompt for GitHub token
- Create qv.db with the required schema`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
