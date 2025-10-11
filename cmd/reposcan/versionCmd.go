package reposcan

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "dev" // overridden at build time
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of reposcan",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("reposcan %s\n", version)

		// If built with Go 1.18+ and no ldflags, show info from the build metadata
		if info, ok := debug.ReadBuildInfo(); ok && version == "dev" {
			fmt.Printf("module: %s\n", info.Main.Path)
			fmt.Printf("go: %s\n", info.GoVersion)
		}

		if date != "unknown" {
			fmt.Printf("built: %s\n", date)
		}
	},
}
