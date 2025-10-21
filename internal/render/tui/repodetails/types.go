package repodetails

import (
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Model struct {
	repoState *report.RepoState
	theme     theme.Theme
}
