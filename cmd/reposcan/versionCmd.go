package reposcan

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of reposcan",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("reposcan 1.3.4\n")
	},
}
