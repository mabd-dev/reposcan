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

	keys := []string{}
	descs := []string{}
	for _, kb := range keybindings {
		keys = append(keys, keybindingStyle.Render(kb.Key+" -"))
		descs = append(descs, theme.Styles.PopupText.Render(" "+kb.Description))
	}

	lines := lipgloss.JoinVertical(
		lipgloss.Center,
		theme.Styles.PopupHeader.Render("Keybindings"),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.JoinVertical(
				lipgloss.Right,
				keys...,
			),
			lipgloss.JoinVertical(
				lipgloss.Left,
				descs...,
			),
		),
	)

	return theme.Styles.Popup.Render(lines)
}
