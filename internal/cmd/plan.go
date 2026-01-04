package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("plan command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().BoolVar(&majorFlag, "major", false, "increment major version")
	planCmd.Flags().BoolVar(&patchFlag, "patch", false, "increment patch version")
}
