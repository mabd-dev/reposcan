package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MABD-dev/RepoScan/internal/config"
	"github.com/MABD-dev/RepoScan/internal/gitx"
	"github.com/MABD-dev/RepoScan/internal/scan"
	"github.com/MABD-dev/RepoScan/internal/utils"
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
	paths := config.DefaultPaths()

	// Step 1: Reading config and create default file if not exists
	configs, err := config.CreateOrReadConfigs(paths.ConfigFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	validation := config.Validate(configs)
	if len(validation.Errors) > 0 {
		validation.Print()
		os.Exit(1)
	}

	// Step 2: define cli subcommands
	var roots multiFlag

	flag.Var(&roots, "root", "Root directory to scan. Defaults to $HOME.")
	jsonStdout := flag.Bool("json-stdout", true, "Write resport to stdout in json format")
	only := flag.String("only", "all", "Filter out git repos, options=all|uncommited")
	flag.Parse()

	if len(roots) == 0 {
		roots = configs.Roots
	} else {
		configs.Roots = roots
	}

	if jsonStdout != nil {
		configs.JsonStdOut = *jsonStdout
	}

	if only != nil {
		onlyFilter, err := config.CreateOnlyFilter(*only)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		configs.Only = onlyFilter
	}

	// validate after applied cli commands to config
	validation = config.Validate(configs)
	if len(validation.Errors) > 0 {
		validation.Print()
		os.Exit(1)
	}

	fmt.Printf("Look into roots=%s\n", configs.Roots)

	// Step 3: find git repos at defined configs.Roots
	gitReposPaths, warnings := scan.FindGitRepos(configs.Roots)

	for _, warning := range warnings {
		fmt.Println("warning: " + warning)
	}

	repoStates := make([]report.RepoState, 0, len(gitReposPaths))

	for _, repoPath := range gitReposPaths {
		gitRepo := gitx.CreateGitReposFrom(repoPath)

		uncommitedLines, err := gitRepo.UncommitedFiles()
		if err != nil {
			fmt.Println("Failed to get uncommited files=" + err.Error())
			continue
		}

		repoStates = append(
			repoStates,
			report.RepoState{
				ID:              utils.Hash(gitRepo.Path),
				Path:            gitRepo.Path,
				Repo:            gitRepo.RepoName,
				Branch:          gitRepo.Branch,
				UncommitedFiles: uncommitedLines,
			},
		)
	}

	report := report.ScanReport{
		Version:     configs.Version,
		GeneratedAt: time.Now(),
		RepoStates:  repoStates,
	}

	reportJson, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		fmt.Println("Error convert report to json, message=", err)
		os.Exit(1)
	}

	if configs.JsonStdOut {
		fmt.Println(string(reportJson))
	}

	for _, repoState := range report.RepoStates {
		if len(repoState.UncommitedFiles) > 0 {
			os.Exit(1)
		}
	}

	os.Exit(0)
}
