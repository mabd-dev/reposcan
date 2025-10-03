// Package repostableheader is a Model for repos table. It uses scan report to generate it's
// content and return it to the main Model to render it
package repostableheader

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Header struct {
	repoStatesCount int
	dirtyRepos      int
	Style           Style
}

type Style struct {
	Title    lipgloss.Style
	SubTitle lipgloss.Style
	Dirty    lipgloss.Style
	Clean    lipgloss.Style
}

func (h *Header) SetReport(report report.ScanReport) {
	h.repoStatesCount = len(report.RepoStates)
	h.dirtyRepos = report.DirtyReposCount()

}

func (h *Header) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		h.Style.Title.Render("reposcan"),
		" ",
		h.Style.SubTitle.Render(fmt.Sprintf("• %d repos",
			h.repoStatesCount)),
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
