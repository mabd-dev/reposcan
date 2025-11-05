// Package tui renders scan report in an interactive table
package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/repostableheader"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
	"golang.design/x/clipboard"
)

var (
	totalWidth  int = 100
	totalHeight int = 30

	// width with respect to total window width
	sizeReposTableWidthPercent int = 90

	// height with respect to total window height
	sizeReposTableHeightPercent int = 50
)

func (m Model) Init() tea.Cmd { return nil }

// Render runs a Bubble Tea UI that renders the ScanReport in a table.
func Render(
	r report.ScanReport,
	configs config.Config,
) error {
	colorSchemeName := configs.Output.ColorSchemeName
	colors, err := theme.CreateColors(colorSchemeName)
	if err != nil {
		return err
	}

	theme := theme.Theme{
		Colors: colors,
		Styles: theme.CreateStyles(colors),
	}

	reposTable := repostable.New(
		theme,
		r,
		totalWidth*sizeReposTableWidthPercent/100,
		totalHeight*sizeReposTableHeightPercent/100,
	)

	reposTableHeader := rth.Header{
		Theme: theme,
	}
	reposTableHeader.SetReport(r)

	var repoDetails repodetails.Model
	if len(r.RepoStates) == 0 {
		repoDetails = repodetails.New(nil, theme)
	} else {
		repoDetails = repodetails.New(&r.RepoStates[0], theme)
	}

	m := Model{
		configs:     configs,
		reposTable:  reposTable,
		repoDetails: repoDetails,
		rtHeader:    reposTableHeader,
		alerts:      alerts.New(theme),
		width:       totalWidth,
		height:      totalHeight,
		reposFilter: createRrepoFilter(),
		theme:       theme,
		focusStack:  []FocusState{FocusReposTable},
	}

	err = clipboard.Init()
	if err != nil {
		logger.Warn(err.Error())
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stdout), tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func createRrepoFilter() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Filter by repo/branch name"
	//ti.Focus()
	ti.CharLimit = 156
	ti.Width = 100
	return ti
}
