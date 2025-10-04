package reposcan

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/file"
	"github.com/mabd-dev/reposcan/internal/render/stdout"
	"github.com/mabd-dev/reposcan/internal/render/tui"
	"github.com/mabd-dev/reposcan/internal/scan"
	"github.com/mabd-dev/reposcan/pkg/report"
	"github.com/spf13/cobra"
)

// RootCmd is the root Cobra command implementing the reposcan CLI.
// It parses flags, loads configuration, runs the scan, and renders output.
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

		logger.Init(configs.Debug, paths.LogFileDir)

		return run(configs)
	},
}

// readFlags reads CLI flags from the provided Cobra command and applies them
// to the given config. Flags override values loaded from the config file.
//
// Supported flags:
//   - root (-r)            : repeatable directory roots to scan
//   - dirIgnore (-d)       : repeatable glob patterns to ignore during scan
//   - output (-o)          : output format: json|table|interactive|none
//   - filter (-f)          : repository filter: all|dirty|uncommitted|unpushed|unpulled
//   - json-output-path     : directory to write JSON report files
//   - max-workers (-w)     : number of concurrent git checks
//   - debug (--debug)      : enable/disable debug mode
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
	(*configs).Output.Type = outputFormat

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
	(*configs).Output.JSONPath = jsonOutputPath

	// Read max workers flag
	maxWorkers, err := cmd.Flags().GetInt("max-workers")
	if err != nil {
		return err
	}
	(*configs).MaxWorkers = maxWorkers

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return err
	}
	(*configs).Debug = debug

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

	switch configs.Output.Type {
	case config.OutputJson:
		err := stdout.RenderScanReportAsJson(report)
		if err != nil {
			return err
		}
	case config.OutputTable:
		stdout.RenderScanReportAsTable(report)
	case config.OutputInteractive:
		if err := tui.ShowReportTUI(report); err != nil {
			fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
			os.Exit(1)
		}
	case config.OutputNone:
		// no-output
	}

	trimmedJsonOutputPath := strings.TrimSpace(configs.Output.JSONPath)
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
	case config.OnlyUncommitted:
		if len(repoState.UncommitedFiles) > 0 {
			return true
		}
	case config.OnlyUnpushed:
		if repoState.Ahead > 0 {
			return true
		}
	case config.OnlyUnpulled:
		if repoState.Behind > 0 {
			return true
		}
	}

	return false
}
