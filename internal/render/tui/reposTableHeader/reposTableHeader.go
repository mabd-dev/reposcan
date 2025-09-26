// reposTableHeader is a Model for repos table. It uses scan report to generate it's
// content and return it to the main Model to render it
package reposTableHeader

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/pkg/report"
	"time"
)

type Header struct {
	report report.ScanReport
	Style  Style
}

type Style struct {
	Title    lipgloss.Style
	SubTitle lipgloss.Style
	Dirty    lipgloss.Style
	Clean    lipgloss.Style
}

func (h *Header) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		h.Style.Title.Render("reposcan"),
		" ",
		h.Style.SubTitle.Render(fmt.Sprintf("• %d repos • generated %s",
			len(h.report.RepoStates), h.report.GeneratedAt.Format(time.RFC3339))),
	)

	dirtyRepos := h.report.DirtyReposCount()
	summary := fmt.Sprintf("Total: %d  |  Uncommitted: %d", len(h.report.RepoStates), dirtyRepos)
	if dirtyRepos > 0 {
		summary = h.Style.Dirty.Render(summary)
	} else {
		summary = h.Style.Clean.Render(summary)
	}

	base := lipgloss.JoinVertical(lipgloss.Left,
		header,
		summary,
	)
	return base
}

func (h *Header) SetReport(report report.ScanReport) {
	h.report = report
}
