package utils

import (
	"errors"
	"os"
	"path/filepath"
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

// Write data to file and create all parent folders if needed
func WriteToFile(data []byte, path string) error {
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
