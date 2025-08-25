package main

import (
	"encoding/json"
	"fmt"
	"github.com/MABD-dev/RepoScan/internal/gitx"
	"github.com/MABD-dev/RepoScan/internal/scan"
	"github.com/MABD-dev/RepoScan/pkg/report"
	"time"
)

func main() {
	// TODO: user input -- get list of dir to scan
	// for now assume dirs = ["~/"]
	roots := []string{"/home/mabd/Documents/", "/home/mabd/.config"}

	gitRepos, warnings := scan.FindGitRepos(roots)

	for _, warning := range warnings {
		fmt.Println("warning: " + warning)
	}

	repoStates := make([]report.RepoState, 0, len(gitRepos))
	for _, repoPath := range gitRepos {

		repoName, err := gitx.GitRepoName(repoPath)
		if err != nil {
			fmt.Printf("Failed to get repo name, path=%s, erro=%s\n", repoPath, err.Error())
		}

		branch, err := gitx.GitRepoBranch(repoPath)
		if err != nil {
			fmt.Printf("Failed to get branch name, path=%s, error=%s\n", repoPath, err.Error())
		}

		uncommitedLines, err := gitx.UncommitedFiles(repoPath)
		if err != nil {
			fmt.Println("Failed to get uncommited files=" + err.Error())
			continue
		}

		repoStates = append(
			repoStates,
			report.RepoState{
				Path:            repoPath,
				Repo:            repoName,
				Branch:          branch,
				UncommitedFiles: uncommitedLines,
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
