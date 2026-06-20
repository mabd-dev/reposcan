package config

import (
	"os"
	"path/filepath"
)

// Default configuration locations under the user's home directory.
const (
	DefaultConfigDir  = "/.config/reposcan/"
	DefaultConfigToml = "config.toml"
	DefaultLogFileDir = "/.config/reposcan/logs/"
)

// Paths contains resolved file-system locations for configuration.
type Paths struct {
	ConfigDir      string
	ConfigFilePath string
	LogFileDir     string
}

// DefaultPaths returns the default config directory and file path relative
// to the user's home directory.
func DefaultPaths() Paths {
	return Paths{
		ConfigDir:      DefaultConfigDir,
		ConfigFilePath: DefaultConfigDir + DefaultConfigToml,
		LogFileDir:     DefaultLogFileDir,
	}
}

func (p Paths) GetConfigFullPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(home, p.ConfigFilePath)
	return fullPath, nil
}
