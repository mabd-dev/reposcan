package reposcan

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/file"
	"github.com/mabd-dev/reposcan/internal/render/stdout"
	"github.com/mabd-dev/reposcan/internal/render/tui"
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

		validationResult.Log()

		return run(configs)
	},
}

// readFlags reads CLI flags from the provided Cobra command and applies them
// to the given config. Flags override values loaded from the config file.
//
// Supported flags:
//   - root (-r)					: repeatable directory roots to scan
//   - dirIgnore (-d)       		: repeatable glob patterns to ignore during scan
//   - output (-o)          		: output format: json|table|interactive|none
//   - filter (-f)          		: repository filter: all|dirty|uncommitted|unpushed|unpulled
//   - json-output-path     		: directory to write JSON report files
//   - max-workers (-w)     		: number of concurrent git checks
//   - debug (--debug)      		: enable/disable debug mode
//   - colorscheme (--colorscheme)  : enable/disable debug mode
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

	// Read output colorscheme
	// colorscheme, err := cmd.Flags().GetString("colorscheme")
	// if err != nil {
	// 	return err
	// }
	// (*configs).Output.ColorSchemeName = colorscheme

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
	report := internal.GenerateScanReport(configs)

	switch configs.Output.Type {
	case config.OutputJson:
		err := stdout.RenderScanReportAsJson(report)
		if err != nil {
			return err
		}
	case config.OutputTable:
		stdout.RenderScanReportAsTable(report)
	case config.OutputInteractive:
		if err := tui.Render(report, configs); err != nil {
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
