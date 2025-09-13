package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle    = lipgloss.NewStyle().Bold(true)
	SubtleStyle   = lipgloss.NewStyle().Faint(true)
	HeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	RepoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	BranchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	CleanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	DirtyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	FooterStyle   = lipgloss.NewStyle().Faint(true)
	SectionStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	SelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12"))
)
