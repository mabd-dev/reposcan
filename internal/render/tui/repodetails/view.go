package repodetails

import (
	"fmt"
	"strings"
)

func (m *Model) View() string {

	if m.repoState == nil {
		return ""
	}

	uc := len(m.repoState.UncommitedFiles)

	s := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		m.theme.Styles.Base.Foreground(m.theme.Colors.Muted).Italic(true).Render("\nDetails"),
		fmt.Sprintf("%s %s", s.Render("Repo:"), m.repoState.Repo),
		fmt.Sprintf("%s %s", s.Render("Branch:"), m.repoState.Branch),
		fmt.Sprintf("%s %s", s.Render("Path:"), m.repoState.Path),
	}
	if uc > 0 {
		lines = append(lines, s.Render("Uncommited Files:"))

		files := m.repoState.UncommitedFiles

		maxUncommitedFilesToShow := 3
		trimUncommitedFiles := len(files) > maxUncommitedFilesToShow

		if trimUncommitedFiles {
			files = files[:maxUncommitedFilesToShow]
		}

		for _, f := range files {
			lines = append(lines, "  "+m.theme.Styles.Muted.Render(f))
		}

		if trimUncommitedFiles {
			lines = append(lines, m.theme.Styles.Muted.Render("  ..."))
		}
	}
	return strings.Join(lines, "\n")
}
