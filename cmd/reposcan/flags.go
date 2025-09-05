package reposcan

import (
	"flag"
	cli "github.com/MABD-dev/reposcan/internal/cliFlags"
	"github.com/MABD-dev/reposcan/internal/config"
	"strings"
)

func AddFlagsAndApply(c *config.Config) error {
	var roots cli.MultiFlag
	var outputFormat cli.StringFlag
	var onlyFilter cli.StringFlag
	var jsonOutputPath cli.StringFlag
	var maxWorkers cli.IntFlag

	flag.Var(&roots, "root", "Root directory to scan. Defaults to $HOME.")
	flag.Var(&outputFormat, "output", "Output, option=json|table|none")
	flag.Var(&onlyFilter, "only", "Filter out git repos, options=all|dirty")
	flag.Var(&jsonOutputPath, "json-output-path", "Save scan report into json file")
	flag.Var(&maxWorkers, "max-workers", "number of concurrent git checks")
	flag.Parse()

	if len(roots) != 0 {
		c.Roots = roots
	}

	if outputFormat.IsSet {
		outputFormat, err := config.CreateOutputFormat(outputFormat.Value)
		if err != nil {
			return err
		}
		c.Output = outputFormat
	}

	if onlyFilter.IsSet {
		onlyFilter, err := config.CreateOnlyFilter(onlyFilter.Value)
		if err != nil {
			return err
		}
		c.Only = onlyFilter
	}

	if jsonOutputPath.IsSet {
		path := strings.TrimSpace(jsonOutputPath.Value)
		c.JsonOutputPath = path
	}

	if maxWorkers.IsSet {
		c.MaxWorkers = maxWorkers.Value
	}

	return nil
}
