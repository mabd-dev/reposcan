package reposcan

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MABD-dev/reposcan/internal/config"
	"github.com/MABD-dev/reposcan/internal/gitx"
	"github.com/MABD-dev/reposcan/internal/render/file"
	"github.com/MABD-dev/reposcan/internal/render/stdout"
	"github.com/MABD-dev/reposcan/internal/scan"
	"github.com/MABD-dev/reposcan/pkg/report"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:           "reposcan",
	Short:         "Scan directories for Git repositories and report status",
	Long:          "RepoScan scans one or more root directories for Git repositories and reports uncommitted, ahead/behind status.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Reading config and create default file if not exists
		paths := config.DefaultPaths()
		configs, err := config.CreateOrReadConfigs(paths.ConfigFilePath)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// validationResult configs data are valid
		validationResult := config.Validate(configs)
		if validationResult.IsValid() {
			validationResult.Print()
			return err
		}

		// read flags + override configs data
		err = readFlags(cmd, &configs)
		if err != nil {
			return err
		}

		// validate after overriding existing configs with flags data
		validationResult = config.Validate(configs)
		if validationResult.IsValid() {
			validationResult.Print()
			return fmt.Errorf("invalid configuration after flags")
		}

		return run(configs)
	},
}

// readFlags reads CLI flags from the provided Cobra command and applies them
// to the given config. Flags override values loaded from the config file.
//
// Supported flags:
//   - root (-r)            : repeatable directory roots to scan
//   - dirIgnore (-d)       : repeatable glob patterns to ignore during scan
//   - output (-o)          : output format: json|table|none
//   - filter (-f)          : repository filter: all|dirty
//   - json-output-path     : directory to write JSON report files
//   - max-workers (-w)     : number of concurrent git checks
func readFlags(cmd *cobra.Command, configs *config.Config) error {
	// Read roots flags
	roots, err := cmd.Flags().GetStringArray("root")
	if err != nil {
		return err
	}
	(*configs).Roots = roots

	// Read dirIgnore flags
	dirIgnore, err := cmd.Flags().GetStringArray("dirIgnore")
	if err != nil {
		return err
	}
	if len(dirIgnore) > 0 {
		(*configs).DirIgnore = dirIgnore
	}

	// Read output format flag
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	outputFormat, err := config.CreateOutputFormat(output)
	if err != nil {
		return err
	}
	(*configs).Output = outputFormat

	// Read only-filter flag
	onlyFilterStr, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}
	onlyFilter, err := config.CreateOnlyFilter(onlyFilterStr)
	if err != nil {
		return err
	}
	(*configs).Only = onlyFilter

	// Read json output path flag
	jsonOutputPath, err := cmd.Flags().GetString("json-output-path")
	if err != nil {
		return err
	}
	(*configs).JsonOutputPath = jsonOutputPath

	// Read max workers flag
	maxWorkers, err := cmd.Flags().GetInt("max-workers")
	if err != nil {
		return err
	}
	(*configs).MaxWorkers = maxWorkers

	return nil
}

func run(configs config.Config) error {
	reportWarnings := []string{}

	// Find git repos at defined configs.Roots
	gitReposPaths, warnings := scan.FindGitRepos(configs.Roots, configs.DirIgnore)

	reportWarnings = append(reportWarnings, warnings...)

	repoStates := make([]report.RepoState, 0, len(gitReposPaths))

	allRepoStates, warnings := gitx.GetGitRepoStatesConcurrent(gitReposPaths, configs.MaxWorkers)
	reportWarnings = append(reportWarnings, warnings...)

	// filter repo states based on config OnlyFilter
	for _, repoState := range allRepoStates {
		if filter(configs.Only, repoState) {
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
		err := stdout.RenderScanReportAsJson(report)
		if err != nil {
			return err
		}
	case config.OutputTable:
		stdout.RenderScanReportAsTable(report)
	case config.OutputNone:
		// no-output
	}

	trimmedJsonOutputPath := strings.TrimSpace(configs.JsonOutputPath)
	if len(trimmedJsonOutputPath) > 0 {
		err := file.WriteScanReport(report, trimmedJsonOutputPath)
		if err != nil {
			return err
		}
	}

	for _, repoState := range report.RepoStates {
		if len(repoState.UncommitedFiles) > 0 {
			return errors.New("")
		}
	}

	return nil
}

// Filter repoState based on config only filter
// Returns true if repoState should be in output, false otherwise
func filter(f config.OnlyFilter, repoState report.RepoState) bool {
	switch f {
	case config.OnlyAll:
		return true
	case config.OnlyDirty:
		if repoState.IsDirty() {
			return true
		}
	}

	return false
}
