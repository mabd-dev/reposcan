package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func generateHelpPopup() string {
	lines := lipgloss.JoinVertical(
		lipgloss.Center,
		PopupTitleStyle.Render("Keybindings"),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.JoinVertical(
				lipgloss.Right,
				PopupKeybindingStyle.Render("↑/↓ -"),
				PopupKeybindingStyle.Render("<enter> -"),
				// KeybindingStyle.Render("p -"),
				// KeybindingStyle.Render("P -"),
				// KeybindingStyle.Render("f -"),
				PopupKeybindingStyle.Render("c -"),
				PopupKeybindingStyle.Render("/ -"),
				PopupKeybindingStyle.Render("q -"),
			),
			lipgloss.JoinVertical(
				lipgloss.Left,
				PopupDescriptionStyle.Render(" Navigate up and down (or j/k)"),
				PopupDescriptionStyle.Render(" Open git repository report details"),
				// DescriptionStyle.Render(" Pull changes"),
				// DescriptionStyle.Render(" Push changes"),
				// DescriptionStyle.Render(" Fetch changes"),
				PopupDescriptionStyle.Render(" Copy repo path to clipboard"),
				PopupDescriptionStyle.Render(" Filter by repo/branch name"),
				PopupDescriptionStyle.Render(" Quit"),
			),
		),
	)

	return PopupStyle.Render(lines)
}
