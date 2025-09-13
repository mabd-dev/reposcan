package config

import (
	"errors"
	"strings"
)

// OutputFormat controls how scan results are rendered.
type OutputFormat string

const (
    // OutputJson prints a JSON object representing a ScanReport to stdout.
    OutputJson OutputFormat = "json"

    // OutputTable prints a human-readable table to stdout.
    OutputTable OutputFormat = "table"

    // OutputNone suppresses all stdout output.
    OutputNone OutputFormat = "none"
)

// IsValid reports whether o is a recognized OutputFormat value.
func (o OutputFormat) IsValid() bool {
	switch o {
	case OutputJson, OutputTable, OutputNone:
		return true
	}
	return false
}

// CreateOutputFormat parses s into an OutputFormat, returning an error for
// unrecognized values. Matching is case-insensitive and trims whitespace.
func CreateOutputFormat(s string) (OutputFormat, error) {
	str := strings.ToLower(strings.TrimSpace(s))

	switch str {
	case string(OutputJson):
		return OutputJson, nil
	case string(OutputTable):
		return OutputTable, nil
	case string(OutputNone):
		return OutputNone, nil
	}

	return OutputTable, errors.New(s + " is not valid output format")
}
