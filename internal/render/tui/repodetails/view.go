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
		style.Render("File Changes:"),
	}
	if len(m.repoState.UncommitedFiles) > 0 {
		files := m.repoState.UncommitedFiles

		maxUncommitedFilesToShow := m.height - len(lines) - 1
		trimUncommitedFiles := len(files) > maxUncommitedFilesToShow

		if trimUncommitedFiles {
			files = files[:maxUncommitedFilesToShow]
		}

		for _, f := range files {
			lines = append(lines, "  "+m.theme.Styles.Muted.Render(f))
		}

		if trimUncommitedFiles {
			more := len(m.repoState.UncommitedFiles) - maxUncommitedFilesToShow
			lines = append(lines, m.theme.Styles.Muted.Render("  ... (+"+strconv.Itoa(more)+" more)"))
		}
	} else {
		lines = append(lines, m.theme.Styles.Muted.Render("    no changes"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
