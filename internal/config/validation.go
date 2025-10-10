package config

import (
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/stdout"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/internal/utils"
	"os"
	"slices"
	"strings"
)

// Issue represents a validation warning or error for a configuration field.
type Issue struct {
	Field   string
	Message string
}

// ValidationResult aggregates validation warnings and errors discovered while
// checking a Config value.
type ValidationResult struct {
	Warnings []Issue
	Errors   []Issue
}

// IsValid reports whether the configuration contains any errors.
// It returns true when there is at least one error.
func (v *ValidationResult) IsValid() bool {
	return len(v.Errors) > 0
}

func (v *ValidationResult) Log() {
	for _, warn := range v.Warnings {
		msg := "Configs Validation Warning: Field=" + warn.Field + ", message=" + warn.Message
		logger.Warn(msg)
	}

	for _, err := range v.Errors {
		msg := "Configs Validation Errors: Field=" + err.Field + ", message=" + err.Message
		logger.Error(msg)
	}
}

// Validate checks a Config for common issues such as non-existent roots,
// invalid enum values, and unusable output paths.
func Validate(config Config) ValidationResult {
	warnings := []Issue{}
	errors := []Issue{}

	// validate roots are valid paths
	for _, r := range config.Roots {
		root := os.ExpandEnv(r)
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

	if !config.Output.Type.IsValid() {
		issue := Issue{
			Field:   "Output",
			Message: "'" + string(config.Output.Type) + "' is not a valid OutputFormat",
		}
		errors = append(errors, issue)
	}

	if len(strings.TrimSpace(config.Output.JSONPath)) > 0 {
		outputFileExists, err := utils.DirExists(config.Output.JSONPath)
		if err != nil {
			issue := Issue{
				Field:   "jsonOutputPath",
				Message: "error reading path: '" + config.Output.JSONPath + "' error=" + err.Error(),
			}
			warnings = append(warnings, issue)
		} else if !outputFileExists {
			issue := Issue{
				Field:   "jsonOutputPath",
				Message: "output path '" + config.Output.JSONPath + "' does not exists!",
			}
			warnings = append(warnings, issue)
		}
	}

	colorscheme := strings.ToLower(strings.TrimSpace(config.Output.ColorSchemeName))
	if len(colorscheme) > 0 {
		if !slices.Contains(theme.Schemes, colorscheme) {
			issue := Issue{
				Field:   "output.colorscheme",
				Message: "colorscheme='" + config.Output.ColorSchemeName + "' is invalid",
			}
			warnings = append(warnings, issue)
		}
	}

	return ValidationResult{
		Warnings: warnings,
		Errors:   errors,
	}
}

// Print writes warnings and errors to stdout using the stdout renderer.
func (v ValidationResult) Print() {
	for _, w := range v.Warnings {
		msg := "Confg\tfield=" + w.Field + " , message=" + w.Message + "\n"
		stdout.Warning(msg)
	}

	for _, e := range v.Errors {
		msg := "Config\tfield=" + e.Field + ", message=" + e.Message + "\n"
		stdout.Error(msg)
	}
}
