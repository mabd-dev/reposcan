package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// WriteScanReport writes the given ScanReport as a JSON file into dirPath.
// The file name is derived from the report timestamp. Parent directories are ensured.
func WriteScanReport(
	report report.ScanReport,
	dirPath string,
) error {
	// create folder if it does not exist
	jsonReport, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		msg := "Error convert report to json, message=" + err.Error()
		return errors.New(msg)
	}

	reportFileName := fmt.Sprintf("ScanReport %s.json", report.GeneratedAt.Format("2006-01-02 15-04-05"))
	fullReportPath := filepath.Join(dirPath, reportFileName)

	return utils.WriteToFile(jsonReport, fullReportPath)
}
