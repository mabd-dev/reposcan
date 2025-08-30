package config

import (
	"fmt"
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
			Message: "'" + string(config.Only) + "' is not a valid OnlyFilter value",
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
	if len(v.Warnings) > 0 || len(v.Errors) > 0 {
		fmt.Println("**************** Config ****************")
	}

	for _, w := range v.Warnings {
		fmt.Printf("Confg\tW\tfield=%s, message=%s\n", w.Field, w.Message)
	}

	for _, e := range v.Errors {
		fmt.Printf("Config\tE\tfield=%s, message=%s\n", e.Field, e.Message)
	}

	if len(v.Warnings) > 0 || len(v.Errors) > 0 {
		fmt.Println("***************************************")
	}
}
