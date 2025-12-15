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

	allWorktreeStates      []common.WorktreeState
	filteredWorktreeStates []common.WorktreeState
	filterQuery            string
}
