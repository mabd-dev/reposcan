package stdout

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mabd-dev/reposcan/pkg/report"
)

// RenderScanReportAsJson prints the ScanReport to stdout as pretty-printed JSON.
func RenderScanReportAsJson(r report.ScanReport) error {
	reportJson, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return errors.New("Error convert report to json, message=" + err.Error())
	}

	fmt.Println(string(reportJson))

	return nil
}
