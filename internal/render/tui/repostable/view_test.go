package repostable

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// stubTheme returns a minimal theme sufficient for view tests.
func stubTheme() theme.Theme {
	colors := theme.LipglossScheme{
		Foreground: lipgloss.Color("#ffffff"),
		Accent:     lipgloss.Color("#00ff00"),
		Border:     lipgloss.Color("#444444"),
		BorderActive: lipgloss.Color("#00ff00"),
	}
	styles := theme.Styles{
		Base: lipgloss.NewStyle(),
		Muted: lipgloss.NewStyle(),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.BorderActive),
		BoxMuted: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Border),
	}
	return theme.Theme{Colors: colors, Styles: styles}
}

func TestView_EmptyState_RendersMessage(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{RepoStates: []report.RepoState{}},
		80, // width
		20, // height
	)

	view := m.View()

	if !strings.Contains(view, "No repositories found") {
		t.Errorf("expected empty-state message, got:\n%s", view)
	}

	if !strings.Contains(view, "✨") {
		t.Errorf("expected sparkle emoji in empty-state, got:\n%s", view)
	}
}

func TestView_EmptyState_NoTablePlaceholder(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{RepoStates: []report.RepoState{}},
		80,
		20,
	)

	view := m.View()

	// The old code used to inject an empty table row {{"", "", ""}}.
	// With the fix, the table should not appear at all.
	if strings.Contains(view, "Repo") && strings.Contains(view, "Branch") && strings.Contains(view, "State") {
		t.Errorf("expected NO table headers in empty state, got:\n%s", view)
	}
}

func TestView_WithRepos_RendersTable(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{
				{Repo: "foo", Branch: "main", ID: "foo"},
			},
		},
		80,
		20,
	)

	view := m.View()

	if !strings.Contains(view, "foo") {
		t.Errorf("expected repo name 'foo' in table view, got:\n%s", view)
	}

	if !strings.Contains(view, "Repo") {
		t.Errorf("expected table header 'Repo' in table view, got:\n%s", view)
	}
}

func TestUpdate_EmptyState_DoesNotCrash(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{RepoStates: []report.RepoState{}},
		80,
		20,
	)

	// Simulate a keypress when there are no repos — should not panic.
	_, cmd := m.Update(nil)
	if cmd != nil {
		t.Error("expected nil command for empty-state update")
	}
}

func TestReposCount_Empty_ReturnsZero(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{RepoStates: []report.RepoState{}},
		80,
		20,
	)

	if m.ReposCount() != 0 {
		t.Errorf("expected 0 repos, got %d", m.ReposCount())
	}
}

func TestReposCount_NonEmpty_ReturnsCount(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{
				{Repo: "a", ID: "a"},
				{Repo: "b", ID: "b"},
			},
		},
		80,
		20,
	)

	if m.ReposCount() != 2 {
		t.Errorf("expected 2 repos, got %d", m.ReposCount())
	}
}
