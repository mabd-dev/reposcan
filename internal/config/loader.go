package config

import (
	"fmt"
	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/pelletier/go-toml/v2"
	"os"
	"path/filepath"
)

// WriteToFile serializes config to TOML and writes it to path.
// Parent directories are created if necessary.
func WriteToFile(config Config, path string) error {

	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return utils.WriteToFile(data, path)
}

// Load reads a TOML configuration file from path into conf.
func Load(conf *Config, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, conf)
}

// CreateOrReadConfigs loads configuration from the user's home directory.
// If the config file does not exist, it writes a Defaults() config to disk
// and returns that default configuration.
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
