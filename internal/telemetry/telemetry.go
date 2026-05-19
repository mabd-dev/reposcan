package telemetry

import (
	"fmt"

	"github.com/mabd-dev/reposcan/internal/config"
)

type Telemetry struct {
	UUID   string `json:"uuid"`
	Warned string `json:"warned"`
}

// Send send telemetry usage to analytics service
func Send(
	filter config.OnlyFilter,
	outputFormat config.OutputFormat,
	repoCount int,
	scanDurationMs int,
) {
	fmt.Println("sending telemtry")
}
