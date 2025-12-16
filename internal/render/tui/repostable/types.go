package repostable

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type Model struct {
	width  int
	height int
	theme  theme.Theme

	tbl table.Model

	allRows           []tableRow
	filteredRows      []tableRow
	allWorktreeStates []common.WorktreeState
	filterQuery       string
}

type tableRow struct {
	Repo     string
	Branch   string
	State    string
	IsHeader bool
	WtIndex  int
}
