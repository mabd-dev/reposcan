package render

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/MABD-dev/RepoScan/internal/utils"
	"github.com/MABD-dev/RepoScan/pkg/report"
)

// Prebuilt color formatters (format first, then color)

func RenderScanReport(r report.ScanReport) {
	total := len(r.RepoStates)
	dirty := 0
	for _, rs := range r.RepoStates {
		if len(rs.UncommitedFiles) > 0 {
			dirty++
		}
	}

	// Header
	fmt.Printf("\n\n")
	fmt.Printf("%s\n", BoldS("Repo Scan Report"))
	fmt.Printf("%s %s\n", DimS("Generated at:"), GrayS(r.GeneratedAt.Format(time.RFC3339)))
	if dirty > 0 {
		fmt.Printf("Total repositories: %s  |  With uncommitted changes: %s\n\n",
			BoldS("%d", total), RedS("%d", dirty))
	} else {
		fmt.Printf("Total repositories: %s  |  With uncommitted changes: %s\n\n",
			BoldS("%d", total), GreenS("%d", dirty))
	}

	// Table setup (Path last, not truncated)
	const (
		repoW   = 24
		branchW = 25
		uncommW = 12
	)
	// Header row (use SprintfFunc so widths are applied before coloring)

	if len(r.RepoStates) > 0 {
		fmt.Printf("%s %s %s %s\n",
			CyanBold("%-*s", repoW, "Repo"),
			CyanBold("%-*s", branchW, "Branch"),
			CyanBold("%-*s", uncommW, "Uncommitted"),
			CyanBold("%s", "Path"),
		)
		fmt.Println(strings.Repeat("─", repoW+1+branchW+1+uncommW+1+60-2))
	}

	// Rows
	for _, rs := range r.RepoStates {
		uc := len(rs.UncommitedFiles)
		ucCell := GreenS("%-*d", uncommW, uc)
		if uc > 0 {
			ucCell = RedS("%-*d", uncommW, uc)
		}

		repoCell := fmt.Sprintf("%-*s", repoW, truncateRunes(rs.Repo, repoW))
		branchCell := BlueS("%-*s", branchW, truncateRunes(rs.Branch, branchW))

		fmt.Printf("%s %s %s %s\n",
			repoCell,
			branchCell,
			ucCell,
			rs.Path, // full path, no truncation
		)
	}

	// Details (only if dirty)
	if dirty > 0 {
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
