package config

import (
	"github.com/MABD-dev/RepoScan/internal/render"
	"github.com/MABD-dev/RepoScan/internal/utils"
	"strings"
)

type Issue struct {
	Field   string
	Message string
}

type Validation struct {
	Warnings []Issue
	Errors   []Issue
}

func Validate(config Config) Validation {
	warnings := []Issue{}
	errors := []Issue{}

	// validate roots are valid paths
	for _, root := range config.Roots {
		exists, err := utils.DirExists(root)
		if err != nil {
			issue := Issue{
				Field:   "root",
				Message: "Failed to read " + root + " error=" + err.Error(),
			}
			errors = append(errors, issue)
		} else if !exists {
			issue := Issue{
				Field:   "root",
				Message: "root '" + root + "' does not exist or not a directory",
			}
			errors = append(errors, issue)
		}
	}

	if !config.Only.IsValid() {
		issue := Issue{
			Field:   "Only",
			Message: "'" + string(config.Only) + "' is not a valid OnlyFilter",
		}
		errors = append(errors, issue)
	}

	if !config.Output.IsValid() {
		issue := Issue{
			Field:   "Output",
			Message: "'" + string(config.Output) + "' is not a valid OutputFormat",
		}
		errors = append(errors, issue)
	}

	if len(strings.TrimSpace(config.JsonOutputPath)) > 0 {
		outputFileExists, err := utils.FileExists(config.JsonOutputPath)
		if err != nil {
			issue := Issue{
				Field:   "jsonOutputPath",
				Message: "error reading path: '" + config.JsonOutputPath + "' error=" + err.Error(),
			}
			errors = append(errors, issue)
		} else if !outputFileExists {
			issue := Issue{
				Field:   "jsonOutputPath",
				Message: "output path '" + config.JsonOutputPath + "' does not exists!",
			}
			errors = append(errors, issue)
		}
	}

	return Validation{
		Warnings: warnings,
		Errors:   errors,
	}
}

// Print out warnings and errors to stdout if they exist
func (v Validation) Print() {
	for _, w := range v.Warnings {
		msg := "Confg\tfield=" + w.Field + " , message=" + w.Message + "\n"
		render.Warning(msg)
	}

	for _, e := range v.Errors {
		msg := "Config\tfield=" + e.Field + ", message=" + e.Message + "\n"
		render.Error(msg)
	}
}
