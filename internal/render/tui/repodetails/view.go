package repodetails

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {

	if m.repoState == nil {
		return ""
	}

	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		//m.theme.Styles.Base.Foreground(m.theme.Colors.Muted).Italic(true).Render("Details"),
		fmt.Sprintf("%s %s", style.Render("Path:"), m.repoState.Path),
	}

	if len(m.repoState.UncommitedFiles) > 0 {
		lines = append(lines, style.Render("File Changes:"))
		lines = appendTrimmedList(lines, m.repoState.UncommitedFiles, m.height, func(s string) string {
			return m.theme.Styles.Muted.Render(s)
		})
	}

	if len(m.repoState.OutgoingCommits) > 0 {
		lines = append(lines, style.Render("Outgoing Commits:"))
		lines = appendTrimmedList(lines, m.repoState.OutgoingCommits, m.height, func(s string) string {
			return m.theme.Styles.Muted.Render(s)
		})
	}

	if len(m.repoState.UncommitedFiles) == 0 && len(m.repoState.OutgoingCommits) == 0 {
		lines = append(lines, style.Render("Changes:"))
		lines = append(lines, m.theme.Styles.Muted.Render("    no changes"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func appendTrimmedList(
	lines []string,
	items []string,
	height int,
	render func(string) string,
) []string {
	maxItemsToShow := height - len(lines) - 1
	if maxItemsToShow < 0 {
		maxItemsToShow = 0
	}

	trimmedItems := items
	trimmed := len(items) > maxItemsToShow
	if trimmed {
		trimmedItems = items[:maxItemsToShow]
	}

	for _, item := range trimmedItems {
		lines = append(lines, "  "+render(item))
	}

	if trimmed {
		more := len(items) - maxItemsToShow
		lines = append(lines, render("  ... (+"+strconv.Itoa(more)+" more)"))
	}

	return lines
}
