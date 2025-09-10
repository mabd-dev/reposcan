package reposcan

import (
	"fmt"
	"github.com/mabd-dev/reposcan/internal/config"
	"os"
)

// Execute runs the root Cobra command for the reposcan CLI.
// It exits the process with a non-zero status on error.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	paths := config.DefaultPaths()
	configs, err := config.CreateOrReadConfigs(paths.ConfigFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// validationResult configs data are valid
	validationResult := config.Validate(configs)
	if validationResult.IsValid() {
		validationResult.Print()
		os.Exit(1)
	}

	RootCmd.PersistentFlags().StringArrayP("root", "r", configs.Roots, "Root directory to scan (repeatable). Defaults to $HOME if unset in config.")
	RootCmd.PersistentFlags().StringArrayP("dirIgnore", "d", []string{}, "Glob patterns to ignore during scan (repeatable)")
	RootCmd.PersistentFlags().StringP("output", "o", string(configs.Output), "Output format: json|table|none")
	RootCmd.PersistentFlags().StringP("filter", "f", string(configs.Only), "Repository filter: all|dirty|uncommitted|unpushed|unpulled")
	RootCmd.PersistentFlags().String("json-output-path", configs.JsonOutputPath, "Write scan report JSON files to this directory (optional)")
	RootCmd.PersistentFlags().IntP("max-workers", "w", configs.MaxWorkers, "Number of concurrent git checks")
}
