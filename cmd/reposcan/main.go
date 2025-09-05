package reposcan

import (
	"fmt"
	"github.com/MABD-dev/reposcan/internal/config"
	"github.com/MABD-dev/reposcan/internal/gitx"
	"github.com/MABD-dev/reposcan/internal/render/file"
	"github.com/MABD-dev/reposcan/internal/render/stdout"
	"github.com/MABD-dev/reposcan/internal/scan"
	"github.com/MABD-dev/reposcan/pkg/report"
	"os"
	"strings"
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
	err = AddFlagsAndApply(&configs)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// validate after applied cli commands to config
	validation = config.Validate(configs)
	if len(validation.Errors) > 0 {
		validation.Print()
		os.Exit(1)
	}

	fmt.Printf("Look into roots=%s\n", configs.Roots)

	// Step 3: find git repos at defined configs.Roots
	gitReposPaths, warnings := scan.FindGitRepos(configs.Roots, configs.DirIgnore)

	reportWarnings = append(reportWarnings, warnings...)

	repoStates := make([]report.RepoState, 0, len(gitReposPaths))

	allRepoStates, warnings := gitx.GetGitRepoStatesConcurrent(gitReposPaths, configs.MaxWorkers)
	reportWarnings = append(reportWarnings, warnings...)

	// filter repo states based on config OnlyFilter
	for _, repoState := range allRepoStates {
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
		err = stdout.RenderScanReportAsJson(report)
		if err != nil {
			stdout.Error(err.Error())
			os.Exit(1)
		}
	case config.OutputTable:
		stdout.RenderScanReportAsTable(report)
	case config.OutputNone:
		break
	}

	trimmedJsonOutputPath := strings.TrimSpace(configs.JsonOutputPath)
	if len(trimmedJsonOutputPath) > 0 {
		err = file.WriteScanReport(report, trimmedJsonOutputPath)
		if err != nil {
			stdout.Error(err.Error())
			os.Exit(1)
		}
	}

	for _, repoState := range report.RepoStates {
		if len(repoState.UncommitedFiles) > 0 {
			os.Exit(1)
		}
	}

	os.Exit(0)
}
