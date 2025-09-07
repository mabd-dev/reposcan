package config

import (
	"errors"
	"strings"
)

type OnlyFilter string

const (
    OnlyAll OnlyFilter = "all"

    // Have uncommited files, ahead/behind remote
    OnlyDirty      = "dirty"
    OnlyUncommitted = "uncommitted"
    OnlyUnpushed    = "unpushed"
    OnlyUnpulled    = "unpulled"
)

func (f OnlyFilter) IsValid() bool {
    switch f {
    case OnlyAll, OnlyDirty, OnlyUncommitted, OnlyUnpushed, OnlyUnpulled:
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
    case string(OnlyUncommitted):
        return OnlyUncommitted, nil
    case string(OnlyUnpushed):
        return OnlyUnpushed, nil
    case string(OnlyUnpulled):
        return OnlyUnpulled, nil
    }

    return OnlyAll, errors.New(s + " is not valid only filter")
}
