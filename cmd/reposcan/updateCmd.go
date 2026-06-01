package reposcan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/hashicorp/go-version"
	"github.com/mabd-dev/reposcan/internal"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reposcan",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking for updates...")
		currentVersion := getCurrentVersion()
		latestVersion, err := getLatestVersion()
		if err != nil {
			return fmt.Errorf("something went wrong. error=%v", err.Error())
		}

		needUpdate, err := needToUpdate(currentVersion, latestVersion)
		if err != nil {
			return fmt.Errorf("something went wrong. error=%v", err.Error())
		}

		if !needUpdate {
			fmt.Printf("Already on the latest version (%s)\n", currentVersion)
			return nil
		}

		fmt.Printf("New version available: %s (current: %s)\n", latestVersion, currentVersion)
		fmt.Println("Installing update...")

		if err = update(); err != nil {
			return fmt.Errorf("something went wrong. error=%v", err.Error())
		}

		fmt.Printf("Successfully updated to %s\n", latestVersion)
		return nil
	},
}

func getCurrentVersion() string {
	return internal.VERSION
}

type ghRelease struct {
	TagName string `json:"tag_name"`
}

func getLatestVersion() (string, error) {
	response, err := http.Get("https://api.github.com/repos/mabd-dev/reposcan/releases/latest")
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	var release ghRelease
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName, nil
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
	sh := exec.Command("sh", "-c",
		"curl -fsSL https://raw.githubusercontent.com/mabd-dev/reposcan/main/install.sh | sh")
	sh.Stdout = os.Stdout
	sh.Stderr = os.Stderr
	if err := sh.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}
