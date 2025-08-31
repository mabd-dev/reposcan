package config

import (
	"errors"
	"strings"
)

type OutputFormat string

const (
	// Print json object representing ScanReport
	OutputJson OutputFormat = "json"

	// Print human readable table representing ScanReport
	OutputTable OutputFormat = "table"

	// Print nothing
	OutputNone OutputFormat = "none"
)

func (o OutputFormat) IsValid() bool {
	switch o {
	case OutputJson, OutputTable, OutputNone:
		return true
	}
	return false
}

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
