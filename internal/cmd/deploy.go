package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
