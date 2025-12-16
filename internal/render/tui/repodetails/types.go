package repodetails

import (
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type Model struct {
	height int
	theme  theme.Theme

	worktreeState *common.WorktreeState
}
