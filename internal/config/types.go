package config

import (
	"errors"
	"strings"
)

type OnlyFilter string

const (
	OnlyAll OnlyFilter = "all"

	// Have uncommited files
	OnlyUncommited = "uncommited"
)

func CreateOnlyFilter(s string) (OnlyFilter, error) {
	str := strings.ToLower(strings.TrimSpace(s))

	switch str {
	case string(OnlyAll):
		return OnlyAll, nil
	case string(OnlyUncommited):
		return OnlyUncommited, nil
	}

	return OnlyAll, errors.New(s + " is not valid only filter")
}

func (f OnlyFilter) IsValid() bool {
	switch f {
	case OnlyAll, OnlyUncommited:
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
	JsonStdOut bool `json:"jsonStdOut"`

	Version int `json:"version"`
}

func Defaults() Config {
	return Config{
		Roots:          nil,
		DirIgnore:      nil,
		Only:           OnlyAll,
		JsonOutputPath: "",
		JsonStdOut:     false,
		Version:        1,
	}
}
