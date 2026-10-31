package reposcan

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the root Cobra command for the reposcan CLI.
// It exits the process with a non-zero status on error.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	addRootFlags(RootCmd)

	RootCmd.AddCommand(versionCmd)

	updateCmd.PersistentFlags().StringP("alias", "a", "reposcan", "reposcan binary name")
	RootCmd.AddCommand(updateCmd)
}

// addRootFlags registers the scan flags with zero values. readFlags applies
// only flags the user set, and a zero value keeps "(default ...)" out of
// --help, where it could contradict the user's config.toml.
func addRootFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringArrayP("root", "r", nil, "Root directory to scan (repeatable)")
	flags.StringArrayP("dirIgnore", "d", nil, "Glob patterns to ignore during scan (repeatable)")
	flags.StringP("output", "o", "", "Output format: json|interactive|none")
	flags.StringP("filter", "f", "", "Repository filter: all|dirty|uncommitted|unpushed|unpulled|stash")
	flags.String("json-output-path", "", "Write scan report JSON files to this directory (optional)")
	flags.Int("stale-days", 0, "Only show repos with no local activity for at least N days (0 = disabled). Ignored with --filter unpulled")
	flags.IntP("max-workers", "w", 0, "Number of concurrent git checks")
	flags.Bool("debug", false, "Enable/Disable debug mode")
	flags.Bool("no-telemetry", false, "Enable/Disable sending telemetry")
	// flags.StringP("colorscheme", "", "", "Used only if 'output' is 'interactive'")
}
