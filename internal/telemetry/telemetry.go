// Package telemetry used to send telemetry usage to mixpanel
package telemetry

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/analytics"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/logger"
)

type Telemetry struct {
	UUID   string `json:"uuid"`
	Warned string `json:"warned"`
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
		fmt.Println("Send telemetry")
		return
	}

	// TODO: if !warned -> warn user

	analyticsService := analytics.New(token, debug)

	err := analyticsService.Send("usage", map[string]any{
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"filter":        filter,
		"output_format": outputFormat,
		"repo_count":    repoCount,
		"tool-version":  internal.VERSION,
		"project":       "reposcan",
	})

	if err != nil {
		logger.Error("Failed to send analytics, error")
	}
}
