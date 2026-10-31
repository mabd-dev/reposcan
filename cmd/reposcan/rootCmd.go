package reposcan

import (
	"fmt"
	"strings"

	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/file"
	"github.com/mabd-dev/reposcan/internal/render/stdout"
	"github.com/mabd-dev/reposcan/internal/render/tui"
	"github.com/mabd-dev/reposcan/internal/telemetry"
	"github.com/spf13/cobra"
)

// RootCmd is the root Cobra command implementing the reposcan CLI.
// It loads configuration, applies flags, runs the scan, and renders output.
var RootCmd = &cobra.Command{
	Use:   "reposcan",
	Short: "Scan directories for Git repositories and report status",
	Long: "RepoScan scans one or more root directories for Git repositories and reports uncommitted, ahead/behind status.\n\n" +
		"Flags override the matching keys in ~/.config/reposcan/config.toml.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Reading config and create default file if not exists
		paths := config.DefaultPaths()
		configFullPath, err := paths.GetConfigFullPath()
		if err != nil {
			return err
		}
		configs, err := config.CreateOrReadConfigs(configFullPath)
		if err != nil {
			return err
		}
		// "table" was the default before v1.3.7 and may still be in older config files.
		if configs.Output.Type == config.OutputTable {
			configs.Output.Type = config.OutputInteractive
		}

		// Validate only after flags are applied, so a flag can replace a bad config value.
		err = readFlags(cmd, &configs)
		if err != nil {
			return err
		}

		validationResult := config.Validate(configs)
		if validationResult.HasErrors() {
			validationResult.Print()
			return fmt.Errorf("invalid configuration")
		}

		logger.Init(configs.Debug, paths.LogFileDir)

		validationResult.Log()

		return run(configs)
	},
}

// readFlags applies the flags the user set on the command line to the given
// config. Flags left unset keep the value loaded from the config file.
//
// Supported flags:
//   - root (-r)					: repeatable directory roots to scan
//   - dirIgnore (-d)       		: repeatable glob patterns to ignore during scan
//   - output (-o)          		: output format: json|interactive|none
//   - filter (-f)          		: repository filter: all|dirty|uncommitted|unpushed|unpulled|stash
//   - stale-days				: only show repos with no local activity for at least N days
//   - json-output-path     		: directory to write JSON report files
//   - max-workers (-w)     		: number of concurrent git checks
//   - debug (--debug)      		: enable/disable debug mode
//   - colorscheme (--colorscheme)  : enable/disable debug mode
//   - no-telemetry					: enable/disable sending telemetry data
func readFlags(cmd *cobra.Command, configs *config.Config) error {
	flags := cmd.Flags()

	if flags.Changed("root") {
		roots, err := flags.GetStringArray("root")
		if err != nil {
			return err
		}
		configs.Roots = roots
	}

	if flags.Changed("dirIgnore") {
		dirIgnore, err := flags.GetStringArray("dirIgnore")
		if err != nil {
			return err
		}
		configs.DirIgnore = dirIgnore
	}

	if flags.Changed("output") {
		output, err := flags.GetString("output")
		if err != nil {
			return err
		}
		outputFormat, err := config.CreateOutputFormat(output)
		if err != nil {
			return err
		}
		configs.Output.Type = outputFormat
	}

	if flags.Changed("filter") {
		onlyFilterStr, err := flags.GetString("filter")
		if err != nil {
			return err
		}
		onlyFilter, err := config.CreateOnlyFilter(onlyFilterStr)
		if err != nil {
			return err
		}
		configs.Only = onlyFilter
	}

	if flags.Changed("stale-days") {
		staleDays, err := flags.GetInt("stale-days")
		if err != nil {
			return err
		}
		configs.StaleDays = staleDays
	}

	if flags.Changed("json-output-path") {
		jsonOutputPath, err := flags.GetString("json-output-path")
		if err != nil {
			return err
		}
		configs.Output.JSONPath = jsonOutputPath
	}

	// Read output colorscheme
	// colorscheme, err := cmd.Flags().GetString("colorscheme")
	// if err != nil {
	// 	return err
	// }
	// (*configs).Output.ColorSchemeName = colorscheme

	if flags.Changed("max-workers") {
		maxWorkers, err := flags.GetInt("max-workers")
		if err != nil {
			return err
		}
		configs.MaxWorkers = maxWorkers
	}

	if flags.Changed("debug") {
		debug, err := flags.GetBool("debug")
		if err != nil {
			return err
		}
		configs.Debug = debug
	}

	if flags.Changed("no-telemetry") {
		noTelemetry, err := flags.GetBool("no-telemetry")
		if err != nil {
			return err
		}
		configs.NoTelemetry = noTelemetry
	}

	return nil
}

func run(configs config.Config) error {
	report := internal.GenerateScanReport(configs)

	if !configs.NoTelemetry {
		telemetry.Send(mixpanelToken, configs.Debug, configs.Only, configs.Output.Type, len(report.RepoStates))
	}

	switch configs.Output.Type {
	case config.OutputJson:
		err := stdout.RenderScanReportAsJson(report)
		if err != nil {
			return err
		}
	case config.OutputInteractive:
		if err := tui.Render(report, configs); err != nil {
			return fmt.Errorf("tui error: %w", err)
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

	return nil
}
