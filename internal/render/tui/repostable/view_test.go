package repostable

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// stubTheme returns a minimal theme sufficient for view tests.
func stubTheme() theme.Theme {
	colors := theme.LipglossScheme{
		Foreground:     lipgloss.Color("#ffffff"),
		Accent:         lipgloss.Color("#00ff00"),
		Border:         lipgloss.Color("#444444"),
		BorderActive:   lipgloss.Color("#00ff00"),
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
		TableHeader:    lipgloss.NewStyle(),
		TableRow:       lipgloss.NewStyle(),
		TableSelectedRow: lipgloss.NewStyle(),
	}
	return theme.Theme{Colors: colors, Styles: styles}
}

func TestView_EmptyState_NoReposFound(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{},
			TotalRepos: 0,
		},
		80, // width
		20, // height
	)

	view := m.View()

	if !strings.Contains(view, "No repositories found in the scanned directory") {
		t.Errorf("expected 'no repos found' message, got:\n%s", view)
	}

	if !strings.Contains(view, "🔍") {
		t.Errorf("expected magnifying-glass emoji, got:\n%s", view)
	}
}

func TestView_EmptyState_AllClean(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{},
			TotalRepos: 3, // 3 repos found but all clean / filtered out
		},
		80,
		20,
	)

	view := m.View()

	if !strings.Contains(view, "All repositories are clean") {
		t.Errorf("expected 'all clean' message, got:\n%s", view)
	}

	if !strings.Contains(view, "✨") {
		t.Errorf("expected sparkle emoji, got:\n%s", view)
	}
}

func TestView_EmptyState_NoTablePlaceholder(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{},
			TotalRepos: 0,
		},
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
			TotalRepos: 1,
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
		report.ScanReport{
			RepoStates: []report.RepoState{},
			TotalRepos: 0,
		},
		80,
		20,
	)

	// Simulate a keypress when there are no repos — should not panic.
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, cmd := m.Update(keyMsg)
	if cmd != nil {
		t.Error("expected nil command for empty-state key update")
	}
}

func TestUpdate_EmptyState_WindowSize_PassesThrough(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{},
			TotalRepos: 0,
		},
		80,
		20,
	)

	// Window-size messages should still reach the underlying table.
	wsMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	_, cmd := m.Update(wsMsg)
	if cmd != nil {
		t.Error("expected nil command for window-size in empty state")
	}
}

func TestReposCount_Empty_ReturnsZero(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{},
			TotalRepos: 0,
		},
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
			TotalRepos: 2,
		},
		80,
		20,
	)

	if m.ReposCount() != 2 {
		t.Errorf("expected 2 repos, got %d", m.ReposCount())
	}
}

func TestSetReport_UpdatesTotalRepos(t *testing.T) {
	m := New(
		stubTheme(),
		report.ScanReport{
			RepoStates: []report.RepoState{},
			TotalRepos: 0,
		},
		80,
		20,
	)

	m.SetReport(report.ScanReport{
		RepoStates: []report.RepoState{
			{Repo: "x", ID: "x"},
		},
		TotalRepos: 5,
	})

	if m.totalRepos != 5 {
		t.Errorf("expected totalRepos=5 after SetReport, got %d", m.totalRepos)
	}
}
