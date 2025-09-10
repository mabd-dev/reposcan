package stdout

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/MABD-dev/reposcan/pkg/report"
)

// RenderScanReportAsJson prints the ScanReport to stdout as pretty-printed JSON.
func RenderScanReportAsJson(r report.ScanReport) error {
	reportJson, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return errors.New("Error convert report to json, message=" + err.Error())
	}

	fmt.Println(string(reportJson))

	return nil
}

// RenderScanReportAsTable prints a human-readable table view of the ScanReport
// to stdout. The Path column is printed last and not truncated.
func RenderScanReportAsTable(r report.ScanReport) {
	totalRepos := len(r.RepoStates)
	dirtyRepos := 0
	for _, rs := range r.RepoStates {
		if rs.IsDirty() {
			dirtyRepos++
		}
	}

	Warnings(r.Warnings)
	renderReportHeader(r, totalRepos, dirtyRepos)

	if len(r.RepoStates) > 0 {
		RenderReposTable(r)
	}

	if dirtyRepos > 0 {
		renderDirtyReposDetails(r)
	}
}

func renderReportHeader(r report.ScanReport, totalRepos int, dirtyRepos int) {
	fmt.Printf("\n\n")
	fmt.Printf("%s\n", BoldS("Repo Scan Report"))
	fmt.Printf("%s %s\n", DimS("Generated at:"), GrayS(r.GeneratedAt.Format(time.RFC3339)))
	if dirtyRepos > 0 {
		fmt.Printf("Total repositories: %s  |  Dirty: %s\n\n",
			BoldS("%d", totalRepos), RedS("%d", dirtyRepos))
	} else {
		fmt.Printf("Total repositories: %s  |  Dirty: %s\n\n",
			BoldS("%d", totalRepos), GreenS("%d", dirtyRepos))
	}
}

func renderDirtyReposDetails(r report.ScanReport) {
	fmt.Printf("\n%s\n", CyanBold("Details:"))
	for _, rs := range r.RepoStates {
		if len(rs.UncommitedFiles) == 0 {
			continue
		}
		fmt.Printf("\n%s %s\n%s %s\n",
			MagBold("Repo:"), rs.Repo,
			MagBold("Path:"), rs.Path,
		)
		for _, f := range rs.UncommitedFiles {
			fmt.Printf("  %s\n", GrayS("- %s", f))
		}
	}
}

// truncateRunes truncates to at most n visible runes (avoids breaking alignment with multibyte chars)
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	// leave space for "..."
	if n <= 3 {
		return string([]rune(s)[:n])
	}
	runes := []rune(s)
	return string(runes[:n-3]) + "..."
}
