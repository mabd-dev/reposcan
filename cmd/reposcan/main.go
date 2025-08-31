package reposcan

import (
	"flag"
	"fmt"
	cli "github.com/MABD-dev/reposcan/internal/cliFlags"
	"github.com/MABD-dev/reposcan/internal/config"
	"github.com/MABD-dev/reposcan/internal/gitx"
	"github.com/MABD-dev/reposcan/internal/render"
	"github.com/MABD-dev/reposcan/internal/scan"
	"github.com/MABD-dev/reposcan/internal/utils"
	"github.com/MABD-dev/reposcan/pkg/report"
	"os"
	"time"
)

func Run() {
	reportWarnings := []string{}

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
	var outputFormat cli.StringFlag
	var onlyFilter cli.StringFlag

	flag.Var(&roots, "root", "Root directory to scan. Defaults to $HOME.")
	flag.Var(&outputFormat, "output", "Output, option=json|table|none")
	flag.Var(&onlyFilter, "only", "Filter out git repos, options=all|dirty")
	flag.Parse()

	if len(roots) == 0 {
		roots = configs.Roots
	} else {
		configs.Roots = roots
	}

	if outputFormat.IsSet {
		outputFormat, err := config.CreateOutputFormat(outputFormat.Value)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		configs.Output = outputFormat
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
	reportWarnings = append(reportWarnings, warnings...)

	repoStates := make([]report.RepoState, 0, len(gitReposPaths))

	for _, repoPath := range gitReposPaths {
		gitRepo, warnings := gitx.CreateGitRepoFrom(repoPath)
		reportWarnings = append(reportWarnings, warnings...)

		uncommitedLines, err := gitRepo.UncommitedFiles()
		if err != nil {
			msg := "Failed to get uncommited files=" + err.Error()
			reportWarnings = append(reportWarnings, msg)
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
		Warnings:    reportWarnings,
	}

	switch configs.Output {
	case config.OutputJson:
		err = render.RenderScanReportAsJson(report)
		if err != nil {
			render.Error(err.Error())
			os.Exit(1)
		}
	case config.OutputTable:
		render.RenderScanReportAsTable(report)
	case config.OutputNone:
		break
	}

	for _, repoState := range report.RepoStates {
		if len(repoState.UncommitedFiles) > 0 {
			os.Exit(1)
		}
	}

	os.Exit(0)
}
