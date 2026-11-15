package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// FileExists checks if a file exists at the given path.
// Returns (true, nil) if the file exists,
// (false, nil) if it does not exist,
// or (false, err) if an error other than "not exist" occurs.
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		// file already exists, do nothing
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		// file not found
		return false, nil
	}

	return false, err
}

// DirExists checks if a directory exists at the given path.
// Returns (true, nil) if the directory exists,
// (false, nil) if it does not exist,
// or (false, err) if an error other than "not exist" occurs.
func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// WriteToFile writes data to path, creating parent folders if needed.
// The path is used as-is: callers expand user-supplied paths (ExpandPath)
// first, so resolved paths such as the config file are never re-expanded.
func WriteToFile(data []byte, path string) error {
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// ExpandPath expands a leading "~" (alone, or followed by a path separator)
// to the current user's home directory, then expands environment variables.
//
// Examples:
//
//	ExpandPath("~/Code")      -> "/Users/someone/Code"
//	ExpandPath("$HOME/Code")  -> "/Users/someone/Code"
//	ExpandPath("~bob/Code")   -> "~bob/Code" (other users' homes are not supported)
//	ExpandPath("/tmp/file")   -> "/tmp/file"
//
// The rest of the path only gets OS separators, it is not cleaned, so glob
// patterns and trailing slashes survive. On error the path is returned
// unchanged, with the error from looking up the home directory.
func ExpandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		path = home + filepath.FromSlash(path[1:])
	}
	return os.ExpandEnv(path), nil
}
