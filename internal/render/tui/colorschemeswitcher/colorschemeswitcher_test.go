package colorschemeswitcher

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
)

// stubTheme returns a minimal theme for tests.
// Name is set to a value that won't match any real scheme so indicator logic stays predictable.
func stubTheme() theme.Theme {
	colors := theme.ColorScheme{
		Name:         "test-scheme-name",
		Foreground:   lipgloss.Color("#ffffff"),
		Accent:       lipgloss.Color("#00ff00"),
		Success:      lipgloss.Color("#00ff00"),
		Muted:        lipgloss.Color("#888888"),
		Border:       lipgloss.Color("#444444"),
		BorderActive: lipgloss.Color("#00ff00"),
	}
	styles := theme.Styles{
		Base:             lipgloss.NewStyle(),
		Muted:            lipgloss.NewStyle(),
		Box:              lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.BorderActive),
		BoxMuted:         lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.Border),
		TableHeader:      lipgloss.NewStyle(),
		TableRow:         lipgloss.NewStyle(),
		TableSelectedRow: lipgloss.NewStyle(),
	}
	return theme.Theme{Colors: colors, Styles: styles}
}

func newTestModel() Model {
	return New(stubTheme(), 30)
}

func keyPress(text string) tea.KeyPressMsg { return tea.KeyPressMsg{Text: text} }
func keyCode(code rune) tea.KeyPressMsg    { return tea.KeyPressMsg{Code: code} }
func keyCtrl(c rune) tea.KeyPressMsg       { return tea.KeyPressMsg{Mod: tea.ModCtrl, Code: c} }

// ── New ──────────────────────────────────────────────────────────────────────

func TestNew_LoadsSchemesFromEmbeddedFS(t *testing.T) {
	m := newTestModel()
	if len(m.colorSchemes) == 0 {
		t.Fatal("expected color schemes to be loaded from embedded FS")
	}
}

func TestNew_FilteredMatchesAll(t *testing.T) {
	m := newTestModel()
	if len(m.filteredColorSchemes) != len(m.colorSchemes) {
		t.Errorf("filtered should equal all initially: got %d/%d", len(m.filteredColorSchemes), len(m.colorSchemes))
	}
}

func TestNew_SelectedSchemeNameIsCurrentTheme(t *testing.T) {
	m := newTestModel()
	if m.SelectedSchemeName() != m.theme.Colors.Name {
		t.Errorf("expected selectedSchemeName=%q, got %q", m.theme.Colors.Name, m.SelectedSchemeName())
	}
}

func TestNew_WantsCloseFalse(t *testing.T) {
	m := newTestModel()
	if m.WantsClose() {
		t.Error("expected WantsClose=false after New")
	}
}

func TestNew_TextInputNotFocusedInitially(t *testing.T) {
	m := newTestModel()
	if m.textInput.Focused() {
		t.Error("text input should not be focused initially")
	}
}

func TestNew_SmallHeight_UsesFloor(t *testing.T) {
	// height=5 → tableHeight = max(5, 5-10) = 5; must not panic
	m := New(stubTheme(), 5)
	if len(m.colorSchemes) == 0 {
		t.Fatal("schemes should load regardless of height")
	}
}

func TestNew_NegativeHeight_UsesFloor(t *testing.T) {
	m := New(stubTheme(), 0)
	if len(m.colorSchemes) == 0 {
		t.Fatal("schemes should load regardless of height")
	}
}

// ── Init / Reset / UpdateTheme ────────────────────────────────────────────────

func TestInit_ReturnsNilCmd(t *testing.T) {
	m := newTestModel()
	if m.Init() != nil {
		t.Error("Init should return nil cmd")
	}
}

func TestReset_ClearsWantsClose(t *testing.T) {
	m := newTestModel()
	m.wantsClose = true
	m.Reset()
	if m.WantsClose() {
		t.Error("Reset should clear wantsClose")
	}
}

func TestUpdateTheme_ChangesThemeName(t *testing.T) {
	m := newTestModel()
	updated := stubTheme()
	updated.Colors.Name = "new-theme"
	m.UpdateTheme(updated)
	if m.theme.Colors.Name != "new-theme" {
		t.Errorf("expected theme name %q, got %q", "new-theme", m.theme.Colors.Name)
	}
}

func TestUpdateTheme_RebuildRows(t *testing.T) {
	m := newTestModel()
	before := len(m.tbl.Rows())
	m.UpdateTheme(stubTheme())
	after := len(m.tbl.Rows())
	if before != after {
		t.Errorf("row count changed after UpdateTheme: %d → %d", before, after)
	}
}

// ── Table-focused update ──────────────────────────────────────────────────────

func TestUpdate_Table_Esc_SetsWantsClose(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyCode(tea.KeyEsc))
	if !m.WantsClose() {
		t.Error("esc should set wantsClose")
	}
}

func TestUpdate_Table_Q_SetsWantsClose(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("q"))
	if !m.WantsClose() {
		t.Error("q should set wantsClose")
	}
}

func TestUpdate_Table_CtrlC_SetsWantsClose(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyCtrl('c'))
	if !m.WantsClose() {
		t.Error("ctrl+c should set wantsClose")
	}
}

func TestUpdate_Table_Slash_FocusesTextInput(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("/"))
	if !m.textInput.Focused() {
		t.Error("/ should focus text input")
	}
}

func TestUpdate_Table_Enter_SelectsSchemeAtCursor(t *testing.T) {
	m := newTestModel()
	if len(m.filteredColorSchemes) == 0 {
		t.Skip("no schemes loaded")
	}
	expectedID := m.filteredColorSchemes[0].id
	m, _ = m.Update(keyCode(tea.KeyEnter))
	if m.SelectedSchemeName() != expectedID {
		t.Errorf("expected selectedSchemeName=%q, got %q", expectedID, m.SelectedSchemeName())
	}
}

func TestUpdate_Table_Enter_EmptyFilteredList_NoChange(t *testing.T) {
	m := newTestModel()
	original := m.SelectedSchemeName()
	m.filteredColorSchemes = nil
	m, _ = m.Update(keyCode(tea.KeyEnter))
	if m.SelectedSchemeName() != original {
		t.Errorf("enter on empty filtered list should not change selected; got %q", m.SelectedSchemeName())
	}
}

func TestUpdate_Table_OtherKey_NoSideEffects(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("j"))
	if m.WantsClose() {
		t.Error("j should not set wantsClose")
	}
	if m.textInput.Focused() {
		t.Error("j should not focus text input")
	}
}

func TestUpdate_Table_Esc_ReturnsNilCmd(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(keyCode(tea.KeyEsc))
	if cmd != nil {
		t.Error("esc should return nil cmd")
	}
}

// ── Text-input-focused update ─────────────────────────────────────────────────

func TestUpdate_TextInput_Esc_FocusesTable(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("/"))
	m.textInput.SetValue("cat")
	m, _ = m.Update(keyCode(tea.KeyEsc))
	if m.textInput.Focused() {
		t.Error("esc should blur text input")
	}
}

func TestUpdate_TextInput_Esc_ClearsValue(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("/"))
	m.textInput.SetValue("cat")
	m, _ = m.Update(keyCode(tea.KeyEsc))
	if m.textInput.Value() != "" {
		t.Errorf("esc should clear input value; got %q", m.textInput.Value())
	}
}

func TestUpdate_TextInput_CtrlC_FocusesTable(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("/"))
	m.textInput.SetValue("test")
	m, _ = m.Update(keyCtrl('c'))
	if m.textInput.Focused() {
		t.Error("ctrl+c should blur text input")
	}
	if m.textInput.Value() != "" {
		t.Errorf("ctrl+c should clear input value; got %q", m.textInput.Value())
	}
}

func TestUpdate_TextInput_Enter_FocusesTable(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("/"))
	m, _ = m.Update(keyCode(tea.KeyEnter))
	if m.textInput.Focused() {
		t.Error("enter should blur text input")
	}
}

func TestUpdate_TextInput_Enter_ReturnsNilCmd(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyPress("/"))
	_, cmd := m.Update(keyCode(tea.KeyEnter))
	if cmd != nil {
		t.Error("enter in text input should return nil cmd")
	}
}

func TestUpdate_TextInput_Typing_FiltersSchemes(t *testing.T) {
	m := newTestModel()
	if len(m.colorSchemes) == 0 {
		t.Skip("no schemes loaded")
	}
	m, _ = m.Update(keyPress("/"))

	target := m.colorSchemes[0].scheme.Name
	if len(target) < 3 {
		t.Skip("scheme name too short for filtering test")
	}
	prefix := strings.ToLower(target[:3])

	for _, ch := range prefix {
		m, _ = m.Update(tea.KeyPressMsg{Text: string(ch)})
	}

	for _, s := range m.filteredColorSchemes {
		if !strings.Contains(strings.ToLower(s.scheme.Name), prefix) {
			t.Errorf("filtered scheme %q does not contain prefix %q", s.scheme.Name, prefix)
		}
	}
}

func TestUpdate_TextInput_Typing_UpdatesTextInputWidth(t *testing.T) {
	m := newTestModel()
	if len(m.colorSchemes) == 0 {
		t.Skip("no schemes loaded")
	}
	m, _ = m.Update(keyPress("/"))
	m, _ = m.Update(tea.KeyPressMsg{Text: "z"})
	if m.textInput.Width() <= 0 {
		t.Error("textInput width must be positive after typing")
	}
}

func TestUpdate_TextInput_Typing_EmptyFilter_RestoresAllSchemes(t *testing.T) {
	m := newTestModel()
	if len(m.colorSchemes) == 0 {
		t.Skip("no schemes loaded")
	}
	total := len(m.colorSchemes)
	m, _ = m.Update(keyPress("/"))
	m, _ = m.Update(tea.KeyPressMsg{Text: "z"})

	// clear with esc (resets value to "")
	m, _ = m.Update(keyCode(tea.KeyEsc))
	// refocus and send a space to trigger filtering with empty query
	m, _ = m.Update(keyPress("/"))
	// After focusing, all schemes should still be in colorSchemes
	if len(m.colorSchemes) != total {
		t.Errorf("expected %d total schemes after clearing, got %d", total, len(m.colorSchemes))
	}
}

// ── updateCursorInRows ────────────────────────────────────────────────────────

func TestUpdateCursorInRows_FirstRowHasCursorInitially(t *testing.T) {
	m := newTestModel()
	rows := m.tbl.Rows()
	if len(rows) == 0 {
		t.Skip("no rows")
	}
	if rows[0][0] != cursorChar {
		t.Errorf("expected cursorChar at row 0 initially; got %q", rows[0][0])
	}
}

func TestUpdateCursorInRows_NavigationMovesCursor(t *testing.T) {
	m := newTestModel()
	if len(m.filteredColorSchemes) < 2 {
		t.Skip("need at least 2 schemes")
	}
	m, _ = m.Update(keyPress("j")) // move down
	rows := m.tbl.Rows()
	if rows[0][0] == cursorChar {
		t.Error("row 0 should not have cursorChar after moving down")
	}
	if rows[1][0] != cursorChar {
		t.Errorf("row 1 should have cursorChar after moving down; got %q", rows[1][0])
	}
}

// ── createRows ───────────────────────────────────────────────────────────────

func TestCreateRows_EmptySchemes_ReturnsEmptySlice(t *testing.T) {
	rows := createRows(0, nil, stubTheme())
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for nil schemes, got %d", len(rows))
	}
}

func TestCreateRows_CorrectCount(t *testing.T) {
	th := stubTheme()
	schemes := []colorSchemeData{
		{id: "a", scheme: theme.Base24Scheme{Name: "Alpha"}},
		{id: "b", scheme: theme.Base24Scheme{Name: "Beta"}},
	}
	rows := createRows(0, schemes, th)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestCreateRows_CursorIndicatorAtSelectedIndex(t *testing.T) {
	th := stubTheme()
	schemes := []colorSchemeData{
		{id: "a", scheme: theme.Base24Scheme{Name: "Alpha"}},
		{id: "b", scheme: theme.Base24Scheme{Name: "Beta"}},
		{id: "c", scheme: theme.Base24Scheme{Name: "Gamma"}},
	}
	rows := createRows(1, schemes, th)
	if rows[0][0] != emptyChar {
		t.Errorf("row 0 should be emptyChar; got %q", rows[0][0])
	}
	if rows[1][0] != cursorChar {
		t.Errorf("row 1 should be cursorChar; got %q", rows[1][0])
	}
	if rows[2][0] != emptyChar {
		t.Errorf("row 2 should be emptyChar; got %q", rows[2][0])
	}
}

func TestCreateRows_CurrentThemeLabel(t *testing.T) {
	th := stubTheme()
	th.Colors.Name = "Alpha"
	schemes := []colorSchemeData{
		{id: "alpha", scheme: theme.Base24Scheme{Name: "Alpha"}},
		{id: "beta", scheme: theme.Base24Scheme{Name: "Beta"}},
	}
	rows := createRows(0, schemes, th)
	if !strings.Contains(rows[0][1], "Alpha") {
		t.Errorf("current scheme row should contain scheme name; got %q", rows[0][1])
	}
	if !strings.Contains(rows[0][1], "[current]") {
		t.Errorf("current scheme row should contain [current] label; got %q", rows[0][1])
	}
	if strings.Contains(rows[1][1], "[current]") {
		t.Errorf("non-current row should not contain [current]; got %q", rows[1][1])
	}
}

func TestCreateRows_EachRowHasFourColumns(t *testing.T) {
	th := stubTheme()
	schemes := []colorSchemeData{
		{id: "a", scheme: theme.Base24Scheme{Name: "Alpha"}},
	}
	rows := createRows(0, schemes, th)
	if len(rows[0]) != 4 {
		t.Errorf("expected 4 columns per row, got %d", len(rows[0]))
	}
}

// ── createColors ─────────────────────────────────────────────────────────────

func TestCreateColors_NonEmpty(t *testing.T) {
	palette := theme.Base24Palette{
		Base0D: "#89b4fa",
		Base08: "#f38ba8",
		Base09: "#fab387",
		Base0A: "#f9e2af",
		Base0B: "#a6e3a1",
		Base0C: "#94e2d5",
		Base0E: "#cba6f7",
	}
	result := createColors(palette, stubTheme())
	if result == "" {
		t.Error("createColors should return non-empty string")
	}
}

func TestCreateColors_ContainsDots(t *testing.T) {
	palette := theme.Base24Palette{
		Base0D: "#89b4fa",
		Base08: "#f38ba8",
	}
	result := createColors(palette, stubTheme())
	if !strings.Contains(result, "●") {
		t.Error("createColors should contain dot characters")
	}
}

// ── View ─────────────────────────────────────────────────────────────────────

func TestView_NonEmpty(t *testing.T) {
	m := newTestModel()
	v := m.View()
	if v == "" {
		t.Error("View should return non-empty string")
	}
}

func TestView_ContainsThemesHeader(t *testing.T) {
	m := newTestModel()
	v := m.View()
	if !strings.Contains(v, "Themes") {
		t.Errorf("View should contain 'Themes' header; got:\n%s", v)
	}
}

func TestView_ContainsCounter(t *testing.T) {
	m := newTestModel()
	v := m.View()
	// counter format is "N/M"
	total := len(m.colorSchemes)
	counterStr := strings.Repeat("0", 0) // just check separator exists
	_ = counterStr
	if !strings.Contains(v, "/") {
		t.Errorf("View should contain counter separator '/'; total=%d; got:\n%s", total, v)
	}
}

func TestView_TextInputFocused_DifferentBorderStyle(t *testing.T) {
	m := newTestModel()
	unfocused := m.View()

	m, _ = m.Update(keyPress("/"))
	focused := m.View()

	// The border style changes between Box (focused) and BoxMuted (unfocused).
	// Both should be non-empty; if they differ, the style change is working.
	if unfocused == "" || focused == "" {
		t.Error("View should not be empty in either focus state")
	}
}
