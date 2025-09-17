package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// TODO: move github.com/mabd-dev/internal/render/constants/ module
const (
	RepoW        = 15
	BranchW      = 15
	RemoteStateW = 20 //(uncommited files count + aheadW + behindW + 4 space)
	PathW        = 50
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
)
