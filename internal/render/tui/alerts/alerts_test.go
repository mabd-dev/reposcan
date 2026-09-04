package alerts

import (
	"image/color"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTheme() theme.Theme {
	return theme.Theme{
		Colors: theme.ColorScheme{
			Foreground: lipgloss.Color("#ffffff"),
			Error:      lipgloss.Color("#ff0000"),
			Warning:    lipgloss.Color("#ffff00"),
			Info:       lipgloss.Color("#0000ff"),
			Border:     lipgloss.Color("#888888"),
		},
	}
}

func TestAlertLifecycle(t *testing.T) {
	model := New(testTheme())

	model, cmd := model.Update(AddAlertMsg{Msg: Alert{Title: "First"}})
	require.NotNil(t, cmd)
	model, cmd = model.Update(AddAlertMsg{Msg: Alert{Title: "Second"}})
	assert.Nil(t, cmd)
	require.Len(t, model.alerts, 2)

	for range maxTicks - 1 {
		model, cmd = model.Update(TickMsg{})
	}
	require.Len(t, model.alerts, 2)
	assert.NotNil(t, cmd)

	model, cmd = model.Update(TickMsg{})
	require.Len(t, model.alerts, 1)
	assert.Equal(t, "Second", model.alerts[0].Title)
	assert.NotNil(t, cmd)

	for range maxTicks {
		model, cmd = model.Update(TickMsg{})
	}
	assert.Empty(t, model.alerts)
	assert.Nil(t, cmd)

	model, cmd = model.Update(TickMsg{})
	assert.Empty(t, model.alerts)
	assert.Nil(t, cmd)
}

func TestAlertStatesPositionsAndHidesOverflow(t *testing.T) {
	model := New(testTheme())
	model, _ = model.Update(AddAlertMsg{Msg: Alert{Message: "First"}})
	model, _ = model.Update(AddAlertMsg{Msg: Alert{Message: "Second"}})
	const totalWidth = 40

	unconstrained := model.AlertStates(totalWidth, 100)
	require.Len(t, unconstrained, 2)
	firstHeight := lipgloss.Height(unconstrained[0].AlertView)

	states := model.AlertStates(totalWidth, firstHeight)
	require.Len(t, states, 2)
	for _, state := range states {
		assert.Equal(t, totalWidth, state.X+lipgloss.Width(state.AlertView))
	}
	assert.Equal(t, 0, states[0].Y)
	assert.True(t, states[0].IsVisible)
	assert.Equal(t, firstHeight, states[1].Y)
	assert.False(t, states[1].IsVisible)
}

func TestAlertColorsByType(t *testing.T) {
	model := New(testTheme())
	tests := []struct {
		name          string
		alertType     AlertType
		wantBorder    color.Color
		wantIconColor color.Color
	}{
		{name: "error", alertType: MsgTypeError, wantBorder: model.theme.Colors.Error, wantIconColor: model.theme.Colors.Error},
		{name: "warning", alertType: AlertTypeWarning, wantBorder: model.theme.Colors.Warning, wantIconColor: model.theme.Colors.Warning},
		{name: "info", alertType: AlertTypeInfo, wantBorder: model.theme.Colors.Info, wantIconColor: model.theme.Colors.Info},
		{name: "default", alertType: AlertType("custom"), wantBorder: model.theme.Colors.Border, wantIconColor: model.theme.Colors.Foreground},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			border, iconColor := model.getColors(Alert{Type: tt.alertType})

			assert.Equal(t, tt.wantBorder, border)
			assert.Equal(t, tt.wantIconColor, iconColor)
		})
	}
}

func TestRenderAlertContent(t *testing.T) {
	model := New(testTheme())
	tests := []struct {
		name         string
		alert        Alert
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "title and message",
			alert:        Alert{Type: AlertTypeInfo, Title: "Saved", Message: "Repository updated"},
			wantContains: []string{"ℹ Saved", "Repository updated"},
		},
		{
			name:         "title only",
			alert:        Alert{Type: AlertTypeWarning, Title: "Warning"},
			wantContains: []string{"⚠ Warning"},
		},
		{
			name:         "message only",
			alert:        Alert{Type: MsgTypeError, Message: "Push failed"},
			wantContains: []string{"Push failed"},
			wantMissing:  []string{"✗"},
		},
		{
			name:         "default type",
			alert:        Alert{Type: AlertType("custom"), Title: "Notice"},
			wantContains: []string{"• Notice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := model.renderAlert(tt.alert)

			for _, want := range tt.wantContains {
				assert.Contains(t, view, want)
			}
			for _, unwanted := range tt.wantMissing {
				assert.NotContains(t, view, unwanted)
			}
		})
	}
}

func TestStartTimerReturnsTickMessage(t *testing.T) {
	cmd := startTimer(0 * time.Millisecond)
	require.NotNil(t, cmd)

	assert.IsType(t, TickMsg{}, cmd())
}

// These tests cover a few simple cases so this package stays at 100% test coverage.
// The tests above check the alert behavior that users rely on.
func TestCoverageOnlyPaths(t *testing.T) {
	t.Run("updates the theme", func(t *testing.T) {
		model := New(testTheme())
		updatedTheme := testTheme()
		updatedTheme.Colors.Info = lipgloss.Color("#00ffff")

		model.UpdateTheme(updatedTheme)

		assert.Equal(t, updatedTheme, model.theme)
	})

	t.Run("ignores unknown messages", func(t *testing.T) {
		model := New(testTheme())

		updated, cmd := model.Update(struct{}{})

		assert.Equal(t, model, updated)
		assert.Nil(t, cmd)
	})

	t.Run("returns no states when empty", func(t *testing.T) {
		states := New(testTheme()).AlertStates(80, 24)

		assert.Empty(t, states)
	})
}
