// Package tui renders scan report in an interactive table
package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type reposFilter struct {
	textInput textinput.Model
	show      bool
}

func (rf reposFilter) IsVisible() bool {
	return rf.show && rf.textInput.Focused()
}

type Model struct {
	reposTable        repostable.Table
	rtHeader          rth.Header
	showDetails       bool
	isPushing         bool
	width             int
	height            int
	reposBeingUpdated []string
	warnings          []string
	showHelp          bool
	reposFilter       reposFilter
	theme             theme.Theme
}

func (m *Model) addWarning(msg string) {
	m.warnings = append(m.warnings, msg)
}

func (m *Model) getFocusedModel() focusedModel {
	var focusedModel focusedModel
	if m.showHelp {
		focusedModel = popupFM{}
	} else if m.reposFilter.IsVisible() {
		focusedModel = reposFilterTextFieldFM{}
	} else {
		focusedModel = reposTableFM{}
	}
	return focusedModel
}

// ShowReportTUI runs a Bubble Tea UI that renders the ScanReport in a table.
func ShowReportTUI(r report.ScanReport, colorSchemeName string) error {
	colors, err := theme.CreateColors(colorSchemeName)
	if err != nil {
		return err
	}

	theme := theme.Theme{
		Colors: colors,
		Styles: theme.CreateStyles(colors),
	}

	reposTable := repostable.Table{
		Theme: theme,
	}
	reposTable.SetReport(r)

	reposTable.InitUI(
		totalWidth*sizeReposTableWidthPercent/100,
		totalHeight*sizeReposTableHeightPercent/100,
	)

	reposTableHeader := rth.Header{
		Theme: theme,
	}
	reposTableHeader.SetReport(r)

	m := Model{
		reposTable:  reposTable,
		rtHeader:    reposTableHeader,
		showDetails: false,
		width:       totalWidth,
		height:      totalHeight,
		warnings:    []string{},
		reposFilter: createRrepoFilter(),
		theme:       theme,
	}

	err = clipboard.Init()
	if err != nil {
		m.warnings = append(m.warnings, err.Error())
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stdout), tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func createRrepoFilter() reposFilter {
	ti := textinput.New()
	ti.Placeholder = "Filter by repo/branch name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 100
	return reposFilter{
		textInput: ti,
		show:      false,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.getFocusedModel().update(m, msg)
}

func (m Model) View() string {
	body := m.reposTable.View()

	if m.reposFilter.show {
		focused := m.reposFilter.textInput.Focused()
		textfieldStr := m.theme.Styles.BoxFor(focused).
			Foreground(m.theme.Colors.Foreground).
			Render(m.reposFilter.textInput.View())

		body = lipgloss.JoinVertical(lipgloss.Top, body, textfieldStr)
	}

	if m.showDetails {
		body = lipgloss.JoinVertical(lipgloss.Left, body, m.detailsView())
	}

	header := m.rtHeader.View()
	footer := m.generateFooter()

	// Calculate heights
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	availableHeight := m.height - headerHeight - footerHeight

	body = lipgloss.NewStyle().
		Height(availableHeight).
		MaxHeight(availableHeight).
		Render(body)

	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		body,
		footer,
	)

	view = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		// Background(m.theme.Colors.Background).
		Render(view)

	if m.showHelp {
		helpView := generateHelpPopup(m.theme)

		view = PlaceOverlayWithPosition(
			OverlayPositionCenter,
			m.width, m.height,
			helpView, view,
			true,
			WithWhitespaceChars(" "), // fill empty space
		)
	}

	return view
}

func (m Model) detailsView() string {
	rs := m.reposTable.GetCurrentRepoState()
	if rs == nil {
		return ""
	}

	uc := len(rs.UncommitedFiles)

	s := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		m.theme.Styles.Base.Foreground(m.theme.Colors.Muted).Italic(true).Render("\nDetails"),
		fmt.Sprintf("%s %s", s.Render("Repo:"), rs.Repo),
		fmt.Sprintf("%s %s", s.Render("Branch:"), rs.Branch),
		fmt.Sprintf("%s %s", s.Render("Path:"), rs.Path),
	}
	if uc > 0 {
		lines = append(lines, s.Render("Uncommited Files:"))

		files := rs.UncommitedFiles

		maxUncommitedFilesToShow := 3
		trimUncommitedFiles := len(files) > maxUncommitedFilesToShow

		if trimUncommitedFiles {
			files = files[:maxUncommitedFilesToShow]
		}

		for _, f := range files {
			lines = append(lines, "  "+m.theme.Styles.Muted.Render(f))
		}

		if trimUncommitedFiles {
			lines = append(lines, m.theme.Styles.Muted.Render("  ..."))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *Model) generateFooter() string {
	focusedModel := m.getFocusedModel()
	keybindings := focusedModel.keybindings()

	addKeybinding := false
	switch focusedModel.(type) {
	case reposTableFM:
		addKeybinding = true
	}

	if addKeybinding {
		keybindings = append(keybindings, Keybinding{
			Key:         "?",
			Description: "More Keybindings",
			ShortDesc:   "Keybindings",
		})
	}

	kbStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)
	mutedStyle := m.theme.Styles.Muted

	var sb strings.Builder
	for i, kb := range keybindings {
		sb.WriteString(mutedStyle.Render(kb.ShortDesc))
		sb.WriteString(mutedStyle.Render(": "))
		sb.WriteString(kbStyle.Render(kb.Key))

		if i < len(keybindings)-1 {
			sb.WriteString(mutedStyle.Render(" | "))
		}
	}

	return m.theme.Styles.Muted.Render(sb.String())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
