package config

import (
	"os"
)

type Config struct {
	Roots     []string   `json:"roots,omitempty"`
	DirIgnore []string   `json:"dirignore,omitempty"`
	Only      OnlyFilter `json:"only,omitempty"`

	// Write report json to path, ignored if empty
	JsonOutputPath string `json:"jsonOutputPath,omitempty"`

	// Print json on std out,
	Output OutputFormat `json:"output"`

	Version int `json:"version"`
}

func Defaults() Config {
	home, err := os.UserHomeDir()

	var roots []string = nil
	if err == nil {
		roots = []string{home}
	}

	return Config{
		Roots:          roots,
		DirIgnore:      nil,
		Only:           OnlyDirty,
		JsonOutputPath: "",
		Output:         OutputTable,
		Version:        1,
	}
}
