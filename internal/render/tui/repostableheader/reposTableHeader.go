// Package repostableheader is a Model for repos table. It uses scan report to generate it's
// content and return it to the main Model to render it
package repostableheader

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Header struct {
	repoStatesCount int
	dirtyRepos      int
	Theme           theme.Theme
}

func (h *Header) SetReport(report report.ScanReport) {
	h.repoStatesCount = len(report.RepoStates)
	h.dirtyRepos = report.DirtyReposCount()

}

func (h *Header) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		h.Theme.Styles.Base.Render("reposcan"),
		" ",
		h.Theme.Styles.Base.
			Foreground(h.Theme.Colors.Muted).
			Render(fmt.Sprintf("• %d repos", h.repoStatesCount)),
	)

	// summary := fmt.Sprintf("Total: %d  |  Uncommitted: %d", h.repoStatesCount, h.dirtyRepos)
	// if h.dirtyRepos > 0 {
	// 	summary = h.Style.Dirty.Render(summary)
	// } else {
	// 	summary = h.Style.Clean.Render(summary)
	// }

	base := lipgloss.JoinVertical(lipgloss.Left,
		header,
		// summary,
	)
	return base
}
