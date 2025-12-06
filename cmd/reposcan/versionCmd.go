package reposcan

import (
	"fmt"

	"github.com/mabd-dev/reposcan/internal"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of reposcan",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s", "reposcan "+internal.VERSION+"\n")
	},
}
