package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/repostableheader"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Model struct {
	// Loading stuff
	loading bool
	width   int
	height  int
	theme   theme.Theme

	// configs
	configs           config.Config
	reposBeingUpdated []string

	// Models
	reposTable  repostable.Model
	repoDetails repodetails.Model
	rtHeader    rth.Header
	alerts      alerts.AlertModel
	reposFilter textinput.Model

	focusStack []FocusState
}

func (m Model) IsReposFilterVisible() bool {
	return m.reposFilter.Focused() || len(strings.TrimSpace(m.reposFilter.Value())) != 0
}

type generateReport struct {
	configs config.Config
}

func (g *generateReport) Cmd() tea.Cmd {
	return func() tea.Msg {
		report := internal.GenerateScanReport(g.configs)
		return generateReportResponse{report: report}
	}
}

type generateReportResponse struct {
	report report.ScanReport
}
