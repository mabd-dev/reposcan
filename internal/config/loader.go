package config

import (
	"fmt"
	"os"

	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/pelletier/go-toml/v2"
)

// tomlMarshal is a test seam for forcing marshal failures from writeToFile.
var tomlMarshal = toml.Marshal

// writeToFile serializes config to TOML and writes it to path.
// Parent directories are created if necessary.
func writeToFile(config Config, path string) error {

	data, err := tomlMarshal(config)
	if err != nil {
		return err
	}

	return utils.WriteToFile(data, path)
}

func UpdateConfigs(configs Config, path string) error {
	return writeToFile(configs, path)
}

// SetColorScheme sets output.colorscheme in the config file at path and keeps
// every other key as it is in the file, so values from CLI flags are not saved.
func SetColorScheme(path, schemeID string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	file := map[string]any{}
	if err := toml.Unmarshal(b, &file); err != nil {
		return err
	}
	output, ok := file["output"].(map[string]any)
	if !ok {
		output = map[string]any{}
		file["output"] = output
	}
	output["colorscheme"] = schemeID

	data, err := tomlMarshal(file)
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
func CreateOrReadConfigs(path string) (Config, error) {
	configFileExists, err := utils.FileExists(path)
	if err != nil {
		fmt.Println("Failed to read user config file, error=", err)
		return Config{}, err
	}

	// Start from Defaults() so keys missing from the file keep their default.
	// A key present in the file replaces the default, so `dirIgnore = []` still clears it.
	configs := Defaults()
	if configFileExists {
		Load(&configs, path)
	} else {
		writeToFile(configs, path)
	}

	return configs, nil

}
