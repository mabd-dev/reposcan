package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func generateHelpPopup(theme theme.Theme, keybindings []common.Keybinding) string {

	keybindingStyle := theme.Styles.Base.
		Bold(true).
		Foreground(theme.Colors.Accent)

	// About section
	aboutStyle := theme.Styles.PopupText.Align(lipgloss.Center)
	versionStyle := theme.Styles.PopupText.Align(lipgloss.Center).Foreground(theme.Colors.Accent)

	aboutSection := lipgloss.JoinVertical(
		lipgloss.Center,
		versionStyle.Render("RepoScan v1.3.5"),
		aboutStyle.Render("Developed by MABD"),
		aboutStyle.Render(""),
		aboutStyle.Foreground(theme.Colors.Accent).Render("github.com/mabd-dev/reposcan"),
		aboutStyle.Render("Open Source • Apache-2.0 License"),
	)

	// Keybindings section
	keys := []string{}
	descs := []string{}
	for _, kb := range keybindings {
		keys = append(keys, keybindingStyle.Render(kb.Key+" -"))
		descs = append(descs, theme.Styles.PopupText.Render(" "+kb.Description))
	}

	keybindingsSection := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinVertical(
			lipgloss.Right,
			keys...,
		),
		lipgloss.JoinVertical(
			lipgloss.Left,
			descs...,
		),
	)

	// Combine all sections
	separator := theme.Styles.PopupText.Render("─────────────────────────────────")

	lines := lipgloss.JoinVertical(
		lipgloss.Center,
		theme.Styles.PopupHeader.Render("Help"),
		"",
		aboutSection,
		"",
		separator,
		"",
		theme.Styles.PopupHeader.Align(lipgloss.Center).Render("Keybindings"),
		"",
		keybindingsSection,
	)

	return theme.Styles.Popup.Render(lines)
}
