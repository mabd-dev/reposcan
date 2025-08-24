package main

import (
	"fmt"
	"github.com/MABD-dev/RepoScan/internal/scan"
)

func main() {
	// get list of dir to scan
	// for now assume dirs = ["~/"]
	roots := []string{"/home/mabd/Documents/", "/home/mabd/.config"}

	gitRepos, warnings := scan.FindGitRepos(roots)

	for _, repoPath := range gitRepos {
		fmt.Println("git repo: " + repoPath)
	}

	for _, warning := range warnings {
		fmt.Println("warning: " + warning)
	}

	// scan the dirs
	// write dirs paths
}
