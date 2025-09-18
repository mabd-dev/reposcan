package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func generateHelpPopup(width int, height int) string {
	lines2 := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinVertical(
			lipgloss.Right,
			HeaderStyle.Render("↑/↓"),
			HeaderStyle.Render("Enter"),
			HeaderStyle.Render("p"),
			HeaderStyle.Render("P"),
			HeaderStyle.Render("f"),
			HeaderStyle.Render("c"),
			HeaderStyle.Render("q"),
		),
		lipgloss.JoinVertical(
			lipgloss.Left,
			RepoStyle.Render(" - Navigate up and down (or j/k)"),
			RepoStyle.Render(" - Open git repository report details"),
			RepoStyle.Render(" - Pull changes"),
			RepoStyle.Render(" - Push changes"),
			RepoStyle.Render(" - Fetch changes"),
			RepoStyle.Render(" - Copy repo path to clipboard"),
			RepoStyle.Render(" - Quit"),
		),
	)

	helpBox := PopupStyle.Render(lines2)

	popup := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, helpBox)
	return popup
}
