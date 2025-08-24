package main

import (
	"encoding/json"
	"fmt"
	"github.com/MABD-dev/RepoScan/internal/scan"
	"github.com/MABD-dev/RepoScan/pkg/report"
	"time"
)

func main() {
	// get list of dir to scan
	// for now assume dirs = ["~/"]
	roots := []string{"/home/mabd/Documents/", "/home/mabd/.config"}

	gitRepos, warnings := scan.FindGitRepos(roots)

	for _, warning := range warnings {
		fmt.Println("warning: " + warning)
	}

	repoStates := make([]report.RepoState, 0, len(gitRepos))
	for _, repoPath := range gitRepos {
		fmt.Println("git repo: " + repoPath)
		repoStates = append(
			repoStates,
			report.RepoState{
				Path:       repoPath,
				Repo:       "something",
				Branch:     "Something else",
				Uncommited: false,
			},
		)
	}

	report := report.ScanReport{
		Version:     1,
		GeneratedAt: time.Now(),
		RepoStates:  repoStates,
	}

	reportJson, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		fmt.Println("Error convert report to json, message=", err)
		return
	}

	fmt.Println(string(reportJson))

	// scan the dirs
	// write dirs paths
}
