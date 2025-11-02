package tui

import (
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

type reposFilter struct {
	textInput textinput.Model
	show      bool
}

func (rf reposFilter) IsVisible() bool {
	return rf.show && rf.textInput.Focused()
}

type Model struct {
	loading           bool
	configs           config.Config
	reposTable        repostable.Model
	repoDetails       repodetails.Model
	rtHeader          rth.Header
	alerts            alerts.AlertModel
	isPushing         bool
	width             int
	height            int
	reposBeingUpdated []string
	warnings          []string
	showHelp          bool
	reposFilter       reposFilter
	theme             theme.Theme
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
