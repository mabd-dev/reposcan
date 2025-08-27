package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MABD-dev/RepoScan/internal/gitx"
	"github.com/MABD-dev/RepoScan/internal/scan"
	"github.com/MABD-dev/RepoScan/pkg/report"
	"os"
	"strings"
	"time"
)

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func main() {
	var roots multiFlag
	flag.Var(&roots, "root", "Root directory to scan. Defaults to $HOME.")
	flag.Parse()

	if len(roots) == 0 {
		if home, ok := os.LookupEnv("HOME"); ok {
			roots = append(roots, home)
		} else {
			fmt.Fprintln(os.Stderr, "error: --root not provided and HOME not set")
			os.Exit(1)
		}
	}

	fmt.Printf("Look into roots=%s\n", roots)

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
