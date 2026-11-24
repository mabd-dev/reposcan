package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/theme"
)

var (
	keyCtrlT = tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 't'}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
)

// setTestHome points os.UserHomeDir at dir. It reads HOME on Unix and
// USERPROFILE on Windows, so both are set. An empty dir leaves it undefined.
func setTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

// openThemeSwitcher presses ctrl+t through the production Update path.
func openThemeSwitcher(t *testing.T, m Model) Model {
	t.Helper()
	nm, _ := m.Update(keyCtrlT)
	got := nm.(Model)
	if got.currentFocus() != FocusThemeSwitcher {
		t.Fatalf("focus after ctrl+t = %v, want FocusThemeSwitcher", got.currentFocus())
	}
	return got
}

// pressEnterInThemeSwitcher selects the scheme under the switcher cursor and
// returns the updated model, the returned command and the selected scheme ID.
func pressEnterInThemeSwitcher(t *testing.T, m Model) (Model, tea.Cmd, string) {
	t.Helper()
	nm, cmd := m.Update(keyEnter)
	got := nm.(Model)
	selected := got.colorSchemeSwitcher.SelectedSchemeID()
	if selected == "" || selected == m.theme.Colors.ID {
		t.Fatalf("precondition: enter selected %q, want a scheme other than the current %q", selected, m.theme.Colors.ID)
	}
	return got, cmd, selected
}

func TestThemeSwitcherEnterAppliesAndPersistsSelectedScheme(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)

	m := newTestModel(t)
	originalID := m.theme.Colors.ID
	m.configs.Output.ColorSchemeName = originalID
	m = openThemeSwitcher(t, m)
	before := m.View().Content

	got, cmd, selected := pressEnterInThemeSwitcher(t, m)

	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
	if got.currentFocus() != FocusThemeSwitcher {
		t.Errorf("focus = %v, want the theme switcher to stay open", got.currentFocus())
	}

	wantColors, err := theme.CreateColors(selected)
	if err != nil {
		t.Fatalf("create colors for %q: %v", selected, err)
	}
	if !reflect.DeepEqual(got.theme.Colors, wantColors) {
		t.Errorf("theme colors = %q, want the colors of %q", got.theme.Colors.ID, selected)
	}
	if got.configs.Output.ColorSchemeName != selected {
		t.Errorf("configs colorscheme = %q, want %q", got.configs.Output.ColorSchemeName, selected)
	}

	var saved config.Config
	cfgPath := filepath.Join(home, ".config", "reposcan", "config.toml")
	if err := config.Load(&saved, cfgPath); err != nil {
		t.Fatalf("load saved config: %v", err)
	}
	if saved.Output.ColorSchemeName != selected {
		t.Errorf("saved colorscheme = %q, want %q", saved.Output.ColorSchemeName, selected)
	}

	// Every component is re-themed: the view matches a TUI started with the
	// selected scheme and showing the same switcher.
	after := got.View().Content
	if after == before {
		t.Fatalf("view did not change after switching from %q to %q", originalID, selected)
	}
	reference := openThemeSwitcher(t, newTestModelWithTheme(t, theme.Theme{
		Colors: wantColors,
		Styles: theme.CreateStyles(wantColors),
	}))
	if want := reference.View().Content; after != want {
		t.Errorf("view after switching to %q differs from a TUI started with it\ngot:\n%s\nwant:\n%s", selected, after, want)
	}

	// Alerts shown after the switch use the new theme too.
	alert := alerts.AddAlertMsg{Msg: alerts.Alert{Type: alerts.AlertTypeInfo, Message: "themed alert"}}
	got.alerts, _ = got.alerts.Update(alert)
	reference.alerts, _ = reference.alerts.Update(alert)
	if gotView, want := got.View().Content, reference.View().Content; gotView != want {
		t.Errorf("alert after switching to %q differs from a TUI started with it\ngot:\n%s\nwant:\n%s", selected, gotView, want)
	}
}

func TestThemeSwitcherEnterKeepsThemeWhenConfigCannotBeSaved(t *testing.T) {
	tests := []struct {
		name string
		// setup prepares the home directory and returns it ("" = undefined).
		setup func(t *testing.T) string
	}{
		{
			name:  "home directory undefined",
			setup: func(t *testing.T) string { return "" },
		},
		{
			name: "config directory cannot be created",
			setup: func(t *testing.T) string {
				home := t.TempDir()
				// A regular file where ~/.config should be makes MkdirAll fail.
				if err := os.WriteFile(filepath.Join(home, ".config"), []byte("not a directory"), 0o644); err != nil {
					t.Fatalf("block config dir: %v", err)
				}
				return home
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := tc.setup(t)
			setTestHome(t, home)

			m := newTestModel(t)
			originalID := m.theme.Colors.ID
			m.configs.Output.ColorSchemeName = originalID
			m = openThemeSwitcher(t, m)
			before := m.View().Content

			got, cmd, selected := pressEnterInThemeSwitcher(t, m)

			if cmd != nil {
				t.Errorf("cmd = %v, want nil", cmd)
			}
			if got.currentFocus() != FocusThemeSwitcher {
				t.Errorf("focus = %v, want the theme switcher to stay open", got.currentFocus())
			}
			if got.theme.Colors.ID != originalID {
				t.Errorf("theme = %q, want it to stay %q when %q cannot be saved", got.theme.Colors.ID, originalID, selected)
			}
			if got.configs.Output.ColorSchemeName != originalID {
				t.Errorf("configs colorscheme = %q, want it restored to %q", got.configs.Output.ColorSchemeName, originalID)
			}
			if after := got.View().Content; after != before {
				t.Errorf("view changed although the theme was not applied\nbefore:\n%s\nafter:\n%s", before, after)
			}

			if home != "" {
				info, err := os.Stat(filepath.Join(home, ".config"))
				if err != nil || info.IsDir() {
					t.Errorf("~/.config should still be the blocking file, stat = %v, %v", info, err)
				}
			}
		})
	}
}
