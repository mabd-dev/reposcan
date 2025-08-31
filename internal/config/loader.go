package config

import (
	"fmt"
	"github.com/MABD-dev/reposcan/internal/utils"
	"github.com/pelletier/go-toml/v2"
	"os"
	"path/filepath"
)

// Creates config file at [SourceDefaultConfigPath], then write [config] to it
// If file already exist, do nothing
func WriteToFile(config Config, path string) error {
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func Load(conf *Config, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, conf)
}

// Get config file if exists and load it's data. Or create new one with
// Defaults() function, then save that into newly created file
func CreateOrReadConfigs(configFilePath string) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	filePath := filepath.Join(home, configFilePath)

	configFileExists, err := utils.FileExists(filePath)
	if err != nil {
		fmt.Println("Failed to read user config file, error=", err)
		return Config{}, err
	}

	var configs Config
	if configFileExists {
		Load(&configs, filePath)
		//fmt.Println("loaded config file, data=", configs)
	} else {
		configs = Defaults()
		WriteToFile(configs, filePath)
	}

	return configs, nil

}
