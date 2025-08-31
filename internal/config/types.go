package config

import (
	"errors"
	"os"
	"strings"
)

type OnlyFilter string

const (
	OnlyAll OnlyFilter = "all"

	// Have uncommited files, ahead/behind remote
	OnlyDirty = "dirty"
)

func CreateOnlyFilter(s string) (OnlyFilter, error) {
	str := strings.ToLower(strings.TrimSpace(s))

	switch str {
	case string(OnlyAll):
		return OnlyAll, nil
	case string(OnlyDirty):
		return OnlyDirty, nil
	}

	return OnlyAll, errors.New(s + " is not valid only filter")
}

func (f OnlyFilter) IsValid() bool {
	switch f {
	case OnlyAll, OnlyDirty:
		return true
	}
	return false
}

type Config struct {
	Roots     []string   `json:"roots,omitempty"`
	DirIgnore []string   `json:"dirignore,omitempty"`
	Only      OnlyFilter `json:"only,omitempty"`

	// Write report json to path, ignored if empty
	JsonOutputPath string `json:"jsonOutputPath,omitempty"`

	// Print json on std out,
	PrintStdOut bool `json:"printStdOut"`

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
		PrintStdOut:    true,
		Version:        1,
	}
}
