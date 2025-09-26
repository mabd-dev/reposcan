package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/render/tui/reposTable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/reposTableHeader"
	"github.com/mabd-dev/reposcan/pkg/report"
	"golang.design/x/clipboard"
)

type reposFilter struct {
	textInput textinput.Model
	show      bool
}

func (rf reposFilter) IsVisible() bool {
	return rf.show && rf.textInput.Focused()
}

type Model struct {
	reposTable        reposTable.Table
	rtHeader          rth.Header
	showDetails       bool
	isPushing         bool
	width             int
	height            int
	contentHeight     int
	reposBeingUpdated []string
	warnings          []string
	showHelp          bool
	reposFilter       reposFilter
}

func (m *Model) addWarning(msg string) {
	m.warnings = append(m.warnings, msg)
}

// ShowReportTUI runs a Bubble Tea UI that renders the ScanReport in a table.
func ShowReportTUI(r report.ScanReport) error {
	reposTable := reposTable.Table{
		Style: reposTable.Style{
			Header:      HeaderWithBGStyle,
			SelectedRow: SelectedStyle,
			Cell:        lipgloss.NewStyle(),
		},
	}
	reposTable.SetReport(r)
	reposTable.InitUi()

	reposTableHeader := rth.Header{
		Style: rth.Style{
			Title:    TitleStyle,
			SubTitle: SubtleStyle,
			Dirty:    DirtyStyle,
			Clean:    CleanStyle,
		},
	}
	reposTableHeader.SetReport(r)

	m := Model{
		reposTable:    reposTable,
		rtHeader:      reposTableHeader,
		showDetails:   false,
		width:         100,
		height:        30,
		contentHeight: 18,
		warnings:      []string{},
		reposFilter:   createRrepoFilter(),
	}

	err := clipboard.Init()
	if err != nil {
		m.warnings = append(m.warnings, err.Error())
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stdout), tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func createRrepoFilter() reposFilter {
	ti := textinput.New()
	ti.Placeholder = "Filter by repo name"
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
	var focusedModel focusedModel
	if m.showHelp {
		focusedModel = popupFM{}
	} else if m.reposFilter.IsVisible() {
		focusedModel = reposFilterTextFieldFM{}
	} else {
		focusedModel = reposTableFM{}
	}
	return focusedModel.update(m, msg)
}

func (m Model) View() string {
	if m.showHelp {
		return generateHelpPopup(m.width, m.height)
	}

	header := m.rtHeader.View()
	body := m.reposTable.View()

	if m.reposFilter.show {
		textfieldStr := ReposFilterStyle.Render(m.reposFilter.textInput.View())
		body = lipgloss.JoinVertical(lipgloss.Top, body, textfieldStr)
	}

	if m.showDetails {
		body = lipgloss.JoinVertical(lipgloss.Left, body, m.detailsView())
	}

	// TODO: show most important keybindings here as well
	footer := FooterStyle.Render("↑/↓ to move • ? keybindings")

	var messages strings.Builder
	for _, msg := range m.warnings {
		messages.WriteString(msg)
		messages.WriteString("\n")
	}
	stdMessages := FooterStyle.Render(messages.String())

	base := lipgloss.JoinVertical(lipgloss.Left,
		header,
		body,
		footer,
		stdMessages,
	)

	return base
}

func (m Model) detailsView() string {
	// return ""
	rs := m.reposTable.GetCurrentRepoState()
	if rs == nil {
		return ""
	}

	uc := len(rs.UncommitedFiles)

	// return ""

	lines := []string{
		SectionStyle.Render("\nDetails"),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Repo:"), rs.Repo),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Branch:"), rs.Branch),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Path:"), rs.Path),
	}
	if uc > 0 {
		lines = append(lines, HeaderStyle.Render("Uncommited Files:"))
		for _, f := range rs.UncommitedFiles {
			lines = append(lines, "  "+lipgloss.NewStyle().Faint(true).Render(f))
		}
	}
	return strings.Join(lines, "\n")
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
