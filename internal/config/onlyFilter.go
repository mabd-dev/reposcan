package config

import (
	"errors"
	"strings"
)

// OnlyFilter controls which repositories are included in results based on
// their cleanliness and remote sync state.
type OnlyFilter string

const (
    // OnlyAll includes all repositories, regardless of state.
    OnlyAll OnlyFilter = "all"

    // OnlyDirty includes repositories that have uncommitted changes or are
    // ahead/behind their upstream.
    OnlyDirty       = "dirty"
    // OnlyUncommitted includes repositories with uncommitted files.
    OnlyUncommitted = "uncommitted"
    // OnlyUnpushed includes repositories with local commits not pushed upstream.
    OnlyUnpushed    = "unpushed"
    // OnlyUnpulled includes repositories with upstream commits not yet pulled.
    OnlyUnpulled    = "unpulled"
)

// IsValid reports whether f is a recognized OnlyFilter value.
func (f OnlyFilter) IsValid() bool {
    switch f {
    case OnlyAll, OnlyDirty, OnlyUncommitted, OnlyUnpushed, OnlyUnpulled:
        return true
    }
    return false
}

// CreateOnlyFilter parses s into an OnlyFilter, returning an error for
// unrecognized values. Matching is case-insensitive and trims whitespace.
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
