// Package telemetry used to send telemetry usage to mixpanel
package telemetry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/analytics"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/logger"
)

var (
	toolName                      = "reposcan"
	telemtryFileName              = "telemetry.json"
	userConfigDir                 = os.UserConfigDir
	newAnalyticsService           = analytics.New
	stdout              io.Writer = os.Stdout
)

type Telemetry struct {
	UUID   string `json:"uuid"`
	Warned bool   `json:"warned"`
}

// Send send telemetry usage to analytics service
func Send(
	token string,
	debug bool,
	filter config.OnlyFilter,
	outputFormat config.OutputFormat,
	repoCount int,
	scanDurationMs int,
) {
	isCI := os.Getenv("CI") != ""
	if isCI {
		fmt.Fprintln(stdout, "Send telemetry")
		return
	}

	filePath, err := getTelemetryFilePath()
	if err != nil {
		logger.Error("Failed to get telemetry file path, error=%v", err.Error())
		return
	}
	telemetry, err := getOrCreateTelemetry(filePath)
	if err != nil {
		return
	}

	if !telemetry.Warned {
		fmt.Fprintln(stdout, "reposcan collects anonymous usage telemetry to help improve the tool.")
		fmt.Fprintln(stdout, "No personal data or file paths are collected.")
		fmt.Fprintln(stdout, "To disable: pass --no-telemetry")
		fmt.Fprintln(stdout, "More info: https://github.com/mabd-dev/reposcan#telemetry")

		telemetry.Warned = true
		writeTelemetry(filePath, telemetry)
	}

	analyticsService := newAnalyticsService(token, debug)

	err = analyticsService.Send("usage", map[string]any{
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"filter":        filter,
		"output_format": outputFormat,
		"repo_count":    repoCount,
		"tool-version":  internal.VERSION,
		"project":       toolName,
	})

	if err != nil {
		logger.Error("Failed to send analytics, error")
	}
}

func getOrCreateTelemetry(filePath string) (Telemetry, error) {
	exists, err := fileExists(filePath)
	if err != nil {
		return Telemetry{}, nil
	}

	if exists {
		telemetry, err := readTelemetry(filePath)
		if err != nil {
			logger.Error("Failed to read telemetry, error=%v", err.Error())
			return Telemetry{}, nil
		}
		return telemetry, nil
	}

	telemetry := Telemetry{
		UUID:   uuid.New().String(),
		Warned: false,
	}
	writeTelemetry(filePath, telemetry)
	return telemetry, nil
}

func getTelemetryFilePath() (string, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(configDir, toolName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(dir, telemtryFileName)
	return filePath, nil
}

func readTelemetry(filepath string) (Telemetry, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return Telemetry{}, err
	}

	var telemetry Telemetry
	if err := json.Unmarshal(data, &telemetry); err != nil {
		return Telemetry{}, err
	}

	return telemetry, nil
}

func writeTelemetry(filePath string, telemetry Telemetry) error {
	data, err := json.MarshalIndent(telemetry, "", "    ")
	if err != nil {
		return err
	}

	path, err := expandPath(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// FileExists checks if a file exists at the given path.
// Returns (true, nil) if the file exists,
// (false, nil) if it does not exist,
// or (false, err) if an error other than "not exist" occurs.
func fileExists(path string) (bool, error) {
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

// expandPath expands a filesystem path that may start with '~' into an
// absolute path using the current user's home directory.
//
// Examples:
//
//	expandPath("~/Documents/file.txt")   -> "/Users/someone/Documents/file.txt"
//	expandPath("/tmp/file.txt")          -> "/tmp/file.txt"
//
// Only a leading '~' is expanded. If the path does not start with '~',
// it is returned unchanged.
//
// Returns the expanded absolute path or an error if the home directory
// cannot be determined.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}
