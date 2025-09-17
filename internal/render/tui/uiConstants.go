package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// TODO: move github.com/mabd-dev/internal/render/constants/ module
const (
	RepoW        = 20
	BranchW      = 20
	RemoteStateW = 20 //(uncommited files count + aheadW + behindW + 4 space)
)

var style = lipgloss.NewStyle().
	Align(lipgloss.Left)

var (
	TitleStyle    = style.Bold(true)
	SubtleStyle   = style.Faint(true)
	HeaderStyle   = style.Bold(true).Foreground(lipgloss.Color("6"))
	RepoStyle     = style.Foreground(lipgloss.Color("7"))
	BranchStyle   = style.Foreground(lipgloss.Color("4"))
	CleanStyle    = style.Foreground(lipgloss.Color("2"))
	DirtyStyle    = style.Foreground(lipgloss.Color("1"))
	FooterStyle   = style.Faint(true)
	SectionStyle  = style.Bold(true).Foreground(lipgloss.Color("5"))
	SelectedStyle = style.Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12"))
	DimStyle      = lipgloss.NewStyle().
			Background(lipgloss.Color("#1f1f1f")).
			Foreground(lipgloss.Color("#777")) // soft gray

	PopupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
		// Width(120).
		Align(lipgloss.Left)
)
