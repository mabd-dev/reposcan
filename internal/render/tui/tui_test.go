package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/colorschemeswitcher"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/repostableheader"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func testTheme(t *testing.T) theme.Theme {
	t.Helper()
	colors, err := theme.CreateColors("catppuccin-mocha")
	if err != nil {
		t.Fatalf("create colors: %v", err)
	}
	return theme.Theme{Colors: colors, Styles: theme.CreateStyles(colors)}
}

func sampleScanReport() report.ScanReport {
	return report.ScanReport{
		RepoStates: []report.RepoState{
			{Repo: "alpha", VCSType: "git", Branch: "main", Path: "/tmp/alpha", ID: "a"},
			{Repo: "beta", VCSType: "git", Branch: "dev", Path: "/tmp/beta", ID: "b"},
		},
	}
}

func newTestModel(t *testing.T) Model {
	t.Helper()
	th := testTheme(t)
	rep := sampleScanReport()

	reposTable := repostable.New(th, rep, 100, 30, repostable.Options{ShowVCS: true})
	header := rth.Header{Theme: th}
	header.SetReport(rep, false)
	repoDetails := repodetails.New(&rep.RepoStates[0], th)
	themeSwitcher := colorschemeswitcher.New(th, 30)

	ti := textinput.New()
	ti.Placeholder = "Filter by repo/branch name"
	ti.CharLimit = 156
	ti.SetWidth(100)

	return Model{
		configs:             config.Config{},
		vcsRegistry:         internal.NewVCSRegistry(),
		reposTable:          reposTable,
		repoDetails:         repoDetails,
		rtHeader:            header,
		colorSchemeSwitcher: themeSwitcher,
		alerts:              alerts.New(th),
		reposFilter:         ti,
		theme:               th,
		width:               100,
		height:              30,
		focusStack:          []FocusState{FocusReposTable},
	}
}

func TestUtils(t *testing.T) {
	if got := getRepoIndex([]string{"a", "b", "c"}, "b"); got != 1 {
		t.Errorf("getRepoIndex(b) = %d, want 1", got)
	}
	if got := getRepoIndex([]string{"a"}, "zz"); got != -1 {
		t.Errorf("getRepoIndex(zz) = %d, want -1", got)
	}
	got := deleteRepo([]string{"a", "b", "c"}, 1)
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("deleteRepo = %v, want [a c]", got)
	}
}

func TestShellEscapePath(t *testing.T) {
	if got := shellEscapePath("/a/b"); got != "'/a/b'" {
		t.Errorf("got %q", got)
	}
	got := shellEscapePath("/a/it's/b")
	if !strings.Contains(got, `'\''`) {
		t.Errorf("single quote not escaped: %q", got)
	}
}

func TestFocusModel(t *testing.T) {
	m := newTestModel(t)

	if got := m.currentFocus(); got != FocusReposTable {
		t.Errorf("currentFocus = %v, want FocusReposTable", got)
	}

	m.pushFocus(FocusHelpPopup)
	if got := m.currentFocus(); got != FocusHelpPopup {
		t.Errorf("after push currentFocus = %v", got)
	}

	m.pushFocus(FocusReposFilter)
	m = m.popFocus(true)
	if got := m.currentFocus(); got != FocusHelpPopup {
		t.Errorf("after pop currentFocus = %v, want FocusHelpPopup", got)
	}

	// resetCurrentModel for each focus state
	m2 := newTestModel(t)
	m2.focusStack = []FocusState{FocusReposTable}
	m2.resetCurrentModel()
	m2.focusStack = []FocusState{FocusReposFilter}
	m2.reposFilter.SetValue("x")
	m2.resetCurrentModel()
	if m2.reposFilter.Value() != "" {
		t.Errorf("filter not reset, got %q", m2.reposFilter.Value())
	}
	m2.focusStack = []FocusState{FocusHelpPopup}
	m2.resetCurrentModel()
	m2.focusStack = []FocusState{FocusThemeSwitcher}
	m2.resetCurrentModel()

	// keybindings per focus
	m3 := newTestModel(t)
	m3.focusStack = []FocusState{FocusReposTable}
	if len(m3.keybindings()) == 0 {
		t.Error("expected repos table keybindings")
	}
	m3.focusStack = []FocusState{FocusReposFilter}
	if len(m3.keybindings()) == 0 {
		t.Error("expected filter keybindings")
	}
	m3.focusStack = []FocusState{FocusHelpPopup}
	if len(m3.keybindings()) == 0 {
		t.Error("expected help popup keybindings")
	}
}

func TestCurrentFocusEmptyStack(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = nil
	if got := m.currentFocus(); got != FocusReposTable {
		t.Errorf("currentFocus = %v, want FocusReposTable", got)
	}
}

func TestIsReposFilterVisible(t *testing.T) {
	m := newTestModel(t)
	if m.IsReposFilterVisible() {
		t.Error("unfocused empty filter should not be visible")
	}
	m.reposFilter.Focus()
	if !m.IsReposFilterVisible() {
		t.Error("focused filter should be visible")
	}
	m = newTestModel(t)
	m.reposFilter.SetValue("hello")
	if !m.IsReposFilterVisible() {
		t.Error("filter with value should be visible")
	}
}

func TestGenerateReportCmd(t *testing.T) {
	g := &generateReport{configs: config.Config{}}
	cmd := g.Cmd()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	msg := cmd()
	if _, ok := msg.(generateReportResponse); !ok {
		t.Fatalf("expected generateReportResponse, got %T", msg)
	}
}

func TestGenerateHelpPopupRendersContent(t *testing.T) {
	th := testTheme(t)
	out := generateHelpPopup(th, reposTableKeybindings)
	if !strings.Contains(out, "Help") || !strings.Contains(out, "RepoScan") {
		t.Errorf("help popup missing content: %q", out)
	}
}

type fakeActionProvider struct {
	fetch string
	pull  string
	push  string
	err   error
}

func (f fakeActionProvider) Fetch(path string) (string, error) { return f.fetch, f.err }
func (f fakeActionProvider) Pull(path string) (string, error)  { return f.pull, f.err }
func (f fakeActionProvider) Push(path string) (string, error)  { return f.push, f.err }

type fakeProvider struct {
	state report.RepoState
	warns []string
}

func (f fakeProvider) Type() vcs.Type { return vcs.TypeGit }
func (f fakeProvider) CheckRepoState(path string) (report.RepoState, []string) {
	return f.state, f.warns
}

func TestRunAction(t *testing.T) {
	f := fakeActionProvider{fetch: "f", pull: "p", push: "s"}
	if _, err := runAction(f, vcsActionFetch, "/x"); err != nil {
		t.Errorf("fetch err: %v", err)
	}
	if _, err := runAction(f, vcsActionPull, "/x"); err != nil {
		t.Errorf("pull err: %v", err)
	}
	if _, err := runAction(f, vcsActionPush, "/x"); err != nil {
		t.Errorf("push err: %v", err)
	}
	if _, err := runAction(f, "bogus", "/x"); err == nil {
		t.Error("expected error for unsupported action")
	}
}

func TestUnsupportedVCSActionAlert(t *testing.T) {
	cmd := unsupportedVCSActionAlert(vcsActionPull, "jj")
	msg := cmd()
	add, ok := msg.(alerts.AddAlertMsg)
	if !ok {
		t.Fatalf("expected AddAlertMsg, got %T", msg)
	}
	if add.Msg.Type != alerts.AlertTypeWarning {
		t.Errorf("expected warning type, got %v", add.Msg.Type)
	}

	cmd = unsupportedVCSActionAlert(vcsActionPush, "")
	msg = cmd()
	add, ok = msg.(alerts.AddAlertMsg)
	if !ok {
		t.Fatalf("expected AddAlertMsg, got %T", msg)
	}
	if !strings.Contains(add.Msg.Message, "unknown") {
		t.Errorf("expected 'unknown' in message, got %q", add.Msg.Message)
	}
}

func TestActionProvider(t *testing.T) {
	m := newTestModel(t)
	if _, ok := m.actionProvider(sampleScanReport().RepoStates[0]); !ok {
		t.Error("expected action provider for git repo")
	}

	m.vcsRegistry = nil
	if _, ok := m.actionProvider(sampleScanReport().RepoStates[0]); ok {
		t.Error("expected no action provider when registry is nil")
	}
}

func TestRefreshRepo(t *testing.T) {
	m := newTestModel(t)
	cmd := refreshRepo(m, 0)
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
	msg := cmd()
	if _, ok := msg.(vcsRefreshRepoResultMsg); !ok {
		t.Fatalf("expected vcsRefreshRepoResultMsg, got %T", msg)
	}

	// nil registry -> unsupported alert (covers getRepoStateAt index)
	m2 := newTestModel(t)
	m2.vcsRegistry = nil
	cmd = refreshRepo(m2, 0)
	if cmd == nil {
		t.Fatal("expected command for nil registry")
	}
	msg = cmd()
	if _, ok := msg.(alerts.AddAlertMsg); !ok {
		t.Fatalf("expected AddAlertMsg, got %T", msg)
	}

	// out-of-range index returns nil
	if cmd := refreshRepo(m, 99); cmd != nil {
		t.Error("expected nil command for bad index")
	}
}

func TestRemoveRepoBeingUpdated(t *testing.T) {
	m := newTestModel(t)
	m.reposBeingUpdated = []string{"a", "b", "c"}
	m.removeRepoBeingUpdated("b")
	if len(m.reposBeingUpdated) != 2 || m.reposBeingUpdated[0] != "a" || m.reposBeingUpdated[1] != "c" {
		t.Errorf("unexpected reposBeingUpdated: %v", m.reposBeingUpdated)
	}
	m.removeRepoBeingUpdated("zz") // non-existent index, no-op
}

func TestRunVCSAction(t *testing.T) {
	m := newTestModel(t)
	res, cmd := m.runVCSAction(vcsActionPull)
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	rs, ok := msg.(vcsActionResultMsg)
	if !ok {
		t.Fatalf("expected vcsActionResultMsg, got %T", msg)
	}
	if len(res.reposBeingUpdated) != 1 {
		t.Errorf("expected repo recorded as being updated, got %v", res.reposBeingUpdated)
	}
	if rs.RepoID != "a" {
		t.Errorf("expected repo ID a, got %q", rs.RepoID)
	}
}

func TestRunVCSActionNilRepo(t *testing.T) {
	m := newTestModel(t)
	m.reposTable = repostable.New(testTheme(t), report.ScanReport{}, 100, 30, repostable.Options{ShowVCS: true})
	res, cmd := m.runVCSAction(vcsActionPull)
	if cmd != nil {
		t.Errorf("expected nil command for nil repo, got %v", cmd)
	}
	if got := res.currentFocus(); got != FocusReposTable {
		t.Errorf("unexpected focus %v", got)
	}
}

func TestRunVCSActionUnsupportedProvider(t *testing.T) {
	// A VCS type with no registered ActionProvider surfaces an unsupported alert
	rep := report.ScanReport{RepoStates: []report.RepoState{
		{Repo: "noprov", VCSType: "bogus", Branch: "main", Path: "/tmp/noprov", ID: "n"},
	}}
	m := newTestModel(t)
	m.reposTable = repostable.New(testTheme(t), rep, 100, 30, repostable.Options{ShowVCS: true})
	_, cmd := m.runVCSAction(vcsActionPush)
	if cmd == nil {
		t.Fatal("expected unsupported action alert command")
	}
	msg := cmd()
	if _, ok := msg.(alerts.AddAlertMsg); !ok {
		t.Fatalf("expected AddAlertMsg, got %T", msg)
	}
}

func TestRefreshRepoUnknownVCS(t *testing.T) {
	rep := report.ScanReport{RepoStates: []report.RepoState{
		{Repo: "x", VCSType: "bogus", Branch: "main", Path: "/tmp/x", ID: "x"},
	}}
	m := newTestModel(t)
	m.reposTable = repostable.New(testTheme(t), rep, 100, 30, repostable.Options{ShowVCS: true})
	cmd := refreshRepo(m, 0)
	if cmd == nil {
		t.Fatal("expected unsupported alert command for unknown vcs")
	}
	if _, ok := cmd().(alerts.AddAlertMsg); !ok {
		t.Fatalf("expected AddAlertMsg, got %T", cmd())
	}
}

func TestUpdateFallThroughInvalidFocus(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusState(99)}
	nm, cmd := m.Update(tea.KeyPressMsg{Text: "q"})
	if cmd != nil {
		t.Errorf("expected nil cmd for unknown focus, got %v", cmd)
	}
	if _, ok := nm.(Model); !ok {
		t.Fatalf("expected Model, got %T", nm)
	}
}

func TestUpdateQuitAndFocusChanges(t *testing.T) {
	m := newTestModel(t)

	// q quits
	nm, cmd := m.Update(tea.KeyPressMsg{Text: "q"})
	if cmd == nil {
		t.Error("expected quit command")
	}
	_ = nm

	// "/" pushes focus to filter
	nm, _ = m.Update(tea.KeyPressMsg{Text: "/"})
	m2 := nm.(Model)
	if m2.currentFocus() != FocusReposFilter {
		t.Errorf("expected filter focus, got %v", m2.currentFocus())
	}

	// "ctrl+t" pushes to theme switcher
	nm, _ = newTestModel(t).Update(tea.KeyPressMsg{Text: "ctrl+t"})
	m3 := nm.(Model)
	if m3.currentFocus() != FocusThemeSwitcher {
		t.Errorf("expected theme switcher focus, got %v", m3.currentFocus())
	}

	// "?" pushes to help popup
	nm, _ = newTestModel(t).Update(tea.KeyPressMsg{Text: "?"})
	m4 := nm.(Model)
	if m4.currentFocus() != FocusHelpPopup {
		t.Errorf("expected help popup focus, got %v", m4.currentFocus())
	}
}

func TestUpdateReposTableCopyPath(t *testing.T) {
	m := newTestModel(t)
	nm, cmd := m.updateReposTable(tea.KeyPressMsg{Text: "c"})
	res := nm.(Model)
	if len(res.reposBeingUpdated) != 0 {
		t.Errorf("copy should not mark repo as updating")
	}
	if cmd == nil {
		t.Fatal("expected alert command")
	}
	// execute the command to cover the alert-closure body
	msg := cmd()
	if _, ok := msg.(alerts.AddAlertMsg); !ok {
		t.Fatalf("expected AddAlertMsg, got %T", msg)
	}
}

func TestUpdateReposTableRefresh(t *testing.T) {
	m := newTestModel(t)
	nm, cmd := m.updateReposTable(tea.KeyPressMsg{Text: "r"})
	res := nm.(Model)
	if !res.loading {
		t.Error("expected loading to become true")
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestUpdateReposTableRepoDetailsKeys(t *testing.T) {
	m := newTestModel(t)
	r, _ := m.updateReposTable(tea.KeyPressMsg{Text: "right"})
	nm1 := r.(Model)
	nm2, _ := nm1.updateReposTable(tea.KeyPressMsg{Text: "l"})
	nm3 := nm2.(Model)
	nm3.updateReposTable(tea.KeyPressMsg{Text: "left"})
	nm3.updateReposTable(tea.KeyPressMsg{Text: "h"})
}

func TestUpdateReposFilterEscAndEnter(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusReposTable, FocusReposFilter}
	nm, _ := m.updateReposFilter(tea.KeyPressMsg{Text: "esc"})
	res := nm.(Model)
	if res.currentFocus() != FocusReposTable {
		t.Errorf("esc should return to table focus, got %v", res.currentFocus())
	}

	m = newTestModel(t)
	m.focusStack = []FocusState{FocusReposTable, FocusReposFilter}
	nm, _ = m.updateReposFilter(tea.KeyPressMsg{Text: "enter"})
	res = nm.(Model)
	if res.currentFocus() != FocusReposTable {
		t.Errorf("enter should return to table focus, got %v", res.currentFocus())
	}
}

func TestUpdateReposFilterNonEmptyKeepsFocus(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusReposTable, FocusReposFilter}
	m.reposFilter.SetValue("alp")
	nm, cmd := m.updateReposFilter(tea.WindowSizeMsg{})
	res := nm.(Model)
	if res.currentFocus() != FocusReposFilter {
		t.Errorf("typing should keep filter focus, got %v", res.currentFocus())
	}
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
	if res.reposTable.ReposCount() != 1 {
		t.Errorf("expected 1 filtered repo, got %d", res.reposTable.ReposCount())
	}
}

func TestKeybindingPopup(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusReposTable, FocusHelpPopup}
	nm, _ := m.keybindingPopup(tea.KeyPressMsg{Text: "q"})
	res := nm.(Model)
	if res.currentFocus() != FocusReposTable {
		t.Errorf("expected return to table focus, got %v", res.currentFocus())
	}

	nm, _ = m.keybindingPopup(tea.KeyPressMsg{Text: "esc"})
	res = nm.(Model)
	if res.currentFocus() != FocusReposTable {
		t.Errorf("expected return to table focus on esc, got %v", res.currentFocus())
	}
}

func TestDefaultUpdate(t *testing.T) {
	m := newTestModel(t)

	// WindowSizeMsg
	nm, _ := defaultUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	res := nm.(Model)
	if res.width != 120 || res.height != 40 {
		t.Errorf("window size not updated: %v x %v", res.width, res.height)
	}

	// vcsActionResultMsg with error
	m.reposBeingUpdated = []string{"a"}
	nm, cmd := m.Update(vcsActionResultMsg{RepoID: "a", Err: "boom"})
	res = nm.(Model)
	if len(res.reposBeingUpdated) != 0 {
		t.Errorf("error should remove repo from being updated, got %v", res.reposBeingUpdated)
	}
	if cmd == nil {
		t.Error("expected error alert command")
	}
	if cmd != nil {
		if _, ok := cmd().(alerts.AddAlertMsg); !ok {
			t.Errorf("expected error alert message from cmd")
		}
	}

	// vcsActionResultMsg success
	nm, _ = m.Update(vcsActionResultMsg{RepoID: "a", Index: 0})
	res = nm.(Model)
	if len(res.reposBeingUpdated) != 0 {
		t.Errorf("success should remove repo from being updated")
	}

	// vcsRefreshRepoResultMsg
	nm, _ = m.Update(vcsRefreshRepoResultMsg{index: 0, newRepoState: report.RepoState{}})
	res = nm.(Model)

	// generateReportResponse
	nm, _ = m.Update(generateReportResponse{report: sampleScanReport()})
	res = nm.(Model)
	if res.loading {
		t.Error("loading should be false after report response")
	}
}

func TestUpdateColorSchemeSwitcherWantsClose(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusReposTable, FocusThemeSwitcher}
	// esc on the theme table sets wantsClose -> updateColorSchemeSwitcher pops focus
	sw, _ := m.colorSchemeSwitcher.Update(tea.KeyPressMsg{Text: "esc"})
	m.colorSchemeSwitcher = sw
	if !sw.WantsClose() {
		t.Fatal("expected theme switcher to want close")
	}
	nm, _ := m.updateColorSchemeSwitcher(tea.KeyPressMsg{Text: "esc"})
	res := nm.(Model)
	if res.currentFocus() != FocusReposTable {
		t.Errorf("expected table focus after close, got %v", res.currentFocus())
	}
}

func TestUpdateColorSchemeSwitcherPassthrough(t *testing.T) {
	m := newTestModel(t)
	nm, _ := m.updateColorSchemeSwitcher(tea.WindowSizeMsg{})
	res := nm.(Model)
	if res.colorSchemeSwitcher.WantsClose() {
		t.Error("should not want close")
	}
}

func TestViewLoading(t *testing.T) {
	m := newTestModel(t)
	m.loading = true
	v := m.View()
	if v.Content == "" {
		t.Fatal("expected view content")
	}
	if !v.AltScreen {
		t.Error("expected alt screen during loading")
	}
}

func TestViewRenders(t *testing.T) {
	m := newTestModel(t)
	v := m.View()
	if v.Content == "" {
		t.Fatal("expected view content")
	}
	if v.AltScreen != true {
		t.Error("expected alt screen")
	}
}

func TestGetFooterView(t *testing.T) {
	m := newTestModel(t)
	// filter visible when reposFilter focused
	m.reposFilter.Focus()
	footer := m.getFooterView()
	if len(footer) == 0 {
		t.Error("expected filter footer")
	}

	m2 := newTestModel(t)
	out := m2.generateKeybindingsFooterView()
	if !strings.Contains(out, "Navigate") {
		t.Errorf("expected keybinding footer content, got %q", out)
	}
}

func TestRenderAlerts(t *testing.T) {
	m := newTestModel(t)
	// no alerts -> unchanged
	out := m.renderAlerts("base", []alerts.AlertState{})
	if out != "base" {
		t.Errorf("expected unchanged view, got %q", out)
	}

	m2 := newTestModel(t)
	m2.alerts = alerts.New(testTheme(t))
	am, _ := m2.alerts.Update(alerts.AddAlertMsg{Msg: alerts.Alert{Type: alerts.AlertTypeInfo, Message: "hi"}})
	m2.alerts = am
	states := m2.alerts.AlertStates(100, 30)
	if len(states) != 1 || !states[0].IsVisible {
		t.Errorf("expected one visible alert state, got %v", states)
	}
	out = m2.renderAlerts("base", states)
	if !strings.Contains(out, "hi") {
		t.Errorf("expected alert content in output, got %q", out)
	}
}

func TestKeybindingsVars(t *testing.T) {
	if len(reposTableKeybindings) == 0 {
		t.Error("expected repos table keybindings")
	}
	if len(reposTableFilterKeybindings) == 0 {
		t.Error("expected filter keybindings")
	}
	if len(helpPopupKeybindings) == 0 {
		t.Error("expected help popup keybindings")
	}
}

func TestInitAndCreateRrepoFilter(t *testing.T) {
	m := newTestModel(t)
	if cmd := m.Init(); cmd != nil {
		t.Errorf("expected nil command from Init, got %v", cmd)
	}
	ti := createRrepoFilter()
	if ti.Placeholder != "Filter by repo/branch name" {
		t.Errorf("unexpected placeholder %q", ti.Placeholder)
	}
}

func TestUpdateRoutesToEachFocus(t *testing.T) {
	// Cover every branch of Update's focus switch.
	cases := []struct {
		focus FocusState
		msg   tea.Msg
	}{
		{FocusReposTable, tea.KeyPressMsg{Text: "j"}},
		{FocusReposFilter, tea.KeyPressMsg{Text: "x"}},
		{FocusHelpPopup, tea.KeyPressMsg{Text: "x"}},
		{FocusThemeSwitcher, tea.WindowSizeMsg{}},
	}
	for _, tc := range cases {
		m := newTestModel(t)
		m.focusStack = []FocusState{FocusReposTable, tc.focus}
		m.Update(tc.msg) // must not panic
	}
}

func TestUpdateReposTableNilRepoCopy(t *testing.T) {
	m2 := newTestModel(t)
	m2.reposTable = repostable.New(testTheme(t), report.ScanReport{}, 100, 30, repostable.Options{ShowVCS: true})
	nm, cmd := m2.updateReposTable(tea.KeyPressMsg{Text: "c"})
	res := nm.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd for nil repo copy, got %v", cmd)
	}
	if res.reposBeingUpdated != nil && len(res.reposBeingUpdated) != 0 {
		t.Errorf("copy should not mark repo as updated")
	}
}

func TestUpdateReposTableUnmatchedKey(t *testing.T) {
	m := newTestModel(t)
	// arrow down is not a mapped key; falls through to reposTable.Update
	nm, _ := m.updateReposTable(tea.KeyPressMsg{Text: "down"})
	_ = nm.(Model)
	m2 := newTestModel(t)
	m2.updateReposTable(tea.KeyPressMsg{Text: "ctrl+c"})
}

func TestDefaultUpdateAlertsMsg(t *testing.T) {
	m := newTestModel(t)
	nm, cmd := m.Update(alerts.AddAlertMsg{Msg: alerts.Alert{Type: alerts.AlertTypeInfo, Message: "note"}})
	res := nm.(Model)
	if cmd == nil {
		t.Error("expected alert command")
	}
	_ = res

	// TickMsg also routes to alerts
	nm2, _ := m.Update(alerts.TickMsg{})
	_ = nm2.(Model)
}

func TestKeybindingPopupDefaultUpdate(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusReposTable, FocusHelpPopup}
	// non-keypress msg falls through to defaultUpdate (WindowSizeMsg handled there)
	nm, _ := m.keybindingPopup(tea.WindowSizeMsg{Width: 100, Height: 50})
	res := nm.(Model)
	if res.width != 100 || res.height != 50 {
		t.Errorf("window size not updated in popup: %d x %d", res.width, res.height)
	}
}

func TestViewHelpOverlay(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusHelpPopup}
	v := m.View()
	if v.Content == "" {
		t.Fatal("expected view content")
	}
}

func TestViewThemeSwitcherOverlay(t *testing.T) {
	m := newTestModel(t)
	m.focusStack = []FocusState{FocusThemeSwitcher}
	v := m.View()
	if v.Content == "" {
		t.Fatal("expected view content")
	}
}

func TestVCSConsts(t *testing.T) {
	if vcsActionFetch != "fetch" || vcsActionPull != "pull" || vcsActionPush != "push" {
		t.Error("unexpected vcs action constants")
	}
}

var _ vcs.ActionProvider = fakeActionProvider{}
var _ vcs.Provider = fakeProvider{}
var _ common.Keybinding = common.Keybinding{}
