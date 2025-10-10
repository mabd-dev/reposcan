package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func generateHelpPopup(theme theme.Theme) string {

	keybindingStyle := theme.Styles.Base.
		Bold(true).
		Foreground(theme.Colors.Accent)

	lines := lipgloss.JoinVertical(
		lipgloss.Center,
		theme.Styles.PopupHeader.Render("Keybindings"),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.JoinVertical(
				lipgloss.Right,
				keybindingStyle.Render("↑/↓ -"),
				keybindingStyle.Render("<enter> -"),
				// KeybindingStyle.Render("p -"),
				// KeybindingStyle.Render("P -"),
				// KeybindingStyle.Render("f -"),
				keybindingStyle.Render("c -"),
				keybindingStyle.Render("/ -"),
				keybindingStyle.Render("q -"),
			),
			lipgloss.JoinVertical(
				lipgloss.Left,
				theme.Styles.PopupText.Render(" Navigate up and down (or j/k)"),
				theme.Styles.PopupText.Render(" Open git repository report details"),
				// DescriptionStyle.Render(" Pull changes"),
				// DescriptionStyle.Render(" Push changes"),
				// DescriptionStyle.Render(" Fetch changes"),
				theme.Styles.PopupText.Render(" Copy repo path to clipboard"),
				theme.Styles.PopupText.Render(" Filter by repo/branch name"),
				theme.Styles.PopupText.Render(" Quit"),
			),
		),
	)

	return theme.Styles.Popup.Render(lines)
}
