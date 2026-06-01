package repostable

import (
	"charm.land/bubbles/v2/table"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Model struct {
	width  int
	height int
	theme  theme.Theme

	tbl table.Model

	report        report.ScanReport
	filteredRepos []report.RepoState
	filterQuery   string
}
