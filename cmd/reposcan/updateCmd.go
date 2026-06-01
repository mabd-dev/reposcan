package reposcan

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/mabd-dev/reposcan/internal"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reposcan",
	RunE: func(cmd *cobra.Command, args []string) error {
		currentVersion := getCurrentVersion()
		latestVersion := getLatestVersion()

		needUpdate, err := needToUpdate(currentVersion, latestVersion)
		if err != nil {
			return err
		}

		if needUpdate {
			return update()
		}

		fmt.Println("You are up to date")
		return nil
	},
}

func getCurrentVersion() string {
	return internal.VERSION
}

func getLatestVersion() string {
	return ""
}

func needToUpdate(currentVersionStr string, latestVersionStr string) (bool, error) {
	v1, err := version.NewVersion(currentVersionStr)
	if err != nil {
		return false, err
	}

	v2, err := version.NewVersion(latestVersionStr)
	if err != nil {
		return false, err
	}

	return v1.LessThan(v2), nil
}

func update() error {
	return nil
}
