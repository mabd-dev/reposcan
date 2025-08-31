package main

import (
	"flag"
	"fmt"
	cli "github.com/MABD-dev/RepoScan/internal/cliFlags"
	"github.com/MABD-dev/RepoScan/internal/config"
	"github.com/MABD-dev/RepoScan/internal/gitx"
	"github.com/MABD-dev/RepoScan/internal/render"
	"github.com/MABD-dev/RepoScan/internal/scan"
	"github.com/MABD-dev/RepoScan/internal/utils"
	"github.com/MABD-dev/RepoScan/pkg/report"
	"os"
	"time"
)

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
	var roots cli.MultiFlag
	var printStdout cli.BoolFlag
	var onlyFilter cli.StringFlag

	flag.Var(&roots, "root", "Root directory to scan. Defaults to $HOME.")
	flag.Var(&printStdout, "print-stdout", "Write resport to stdout in table format")
	flag.Var(&onlyFilter, "only", "Filter out git repos, options=all|dirty")
	flag.Parse()

	if len(roots) == 0 {
		roots = configs.Roots
	} else {
		configs.Roots = roots
	}

	if printStdout.IsSet {
		configs.PrintStdOut = printStdout.Value
	}

	if onlyFilter.IsSet {
		onlyFilter, err := config.CreateOnlyFilter(onlyFilter.Value)
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
		render.Warning(warning)
	}

	repoStates := make([]report.RepoState, 0, len(gitReposPaths))

	for _, repoPath := range gitReposPaths {
		gitRepo := gitx.CreateGitRepoFrom(repoPath)

		uncommitedLines, err := gitRepo.UncommitedFiles()
		if err != nil {
			render.Warning("Failed to get uncommited files=" + err.Error())
			continue
		}

		repoState := report.RepoState{
			ID:              utils.Hash(gitRepo.Path),
			Path:            gitRepo.Path,
			Repo:            gitRepo.RepoName,
			Branch:          gitRepo.Branch,
			UncommitedFiles: uncommitedLines,
			Ahead:           gitRepo.Ahead,
			Behind:          gitRepo.Behind,
		}

		if Filter(configs.Only, repoState) {
			repoStates = append(repoStates, repoState)
		}
	}

	report := report.ScanReport{
		Version:     configs.Version,
		GeneratedAt: time.Now(),
		RepoStates:  repoStates,
	}

	if configs.PrintStdOut {
		render.RenderScanReport(report)
	}

	for _, repoState := range report.RepoStates {
		if len(repoState.UncommitedFiles) > 0 {
			os.Exit(1)
		}
	}

	os.Exit(0)
}
