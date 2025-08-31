package config

import (
	"errors"
	"strings"
)

type OnlyFilter string

const (
	OnlyAll OnlyFilter = "all"

	// Have uncommited files, ahead/behind remote
	OnlyDirty = "dirty"
)

func (f OnlyFilter) IsValid() bool {
	switch f {
	case OnlyAll, OnlyDirty:
		return true
	}
	return false
}

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
