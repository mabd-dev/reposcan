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

func TestUpdateIgnoresUnknownMessages(t *testing.T) {
	model := New(testTheme())
	model, _ = model.Update(AddAlertMsg{Msg: Alert{Title: "Keep me"}})

	updated, cmd := model.Update(struct{}{})

	assert.Equal(t, model, updated)
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

func TestAlertStatesReturnsEmptyWithoutAlerts(t *testing.T) {
	states := New(testTheme()).AlertStates(80, 24)

	assert.Empty(t, states)
}

func TestAlertColorsByType(t *testing.T) {
	th := testTheme()
	model := New(th)
	tests := []struct {
		name          string
		alertType     AlertType
		wantBorder    color.Color
		wantIconColor color.Color
	}{
		{name: "error", alertType: MsgTypeError, wantBorder: th.Colors.Error, wantIconColor: th.Colors.Error},
		{name: "warning", alertType: AlertTypeWarning, wantBorder: th.Colors.Warning, wantIconColor: th.Colors.Warning},
		{name: "info", alertType: AlertTypeInfo, wantBorder: th.Colors.Info, wantIconColor: th.Colors.Info},
		{name: "default", alertType: AlertType("custom"), wantBorder: th.Colors.Border, wantIconColor: th.Colors.Foreground},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			border, iconColor := model.getColors(Alert{Type: tt.alertType})

			assert.Equal(t, tt.wantBorder, border)
			assert.Equal(t, tt.wantIconColor, iconColor)
		})
	}
}

func TestUpdateThemeChangesAlertColors(t *testing.T) {
	originalTheme := testTheme()
	model := New(originalTheme)
	updatedTheme := testTheme()
	updatedTheme.Colors.Info = lipgloss.Color("#00ffff")
	updatedTheme.Colors.Border = lipgloss.Color("#123456")
	updatedTheme.Colors.Foreground = lipgloss.Color("#654321")
	require.NotEqual(t, originalTheme.Colors.Info, updatedTheme.Colors.Info)

	_, originalIconColor := model.getColors(Alert{Type: AlertTypeInfo})
	require.Equal(t, originalTheme.Colors.Info, originalIconColor)

	model.UpdateTheme(updatedTheme)

	border, iconColor := model.getColors(Alert{Type: AlertTypeInfo})
	assert.Equal(t, updatedTheme.Colors.Info, border)
	assert.Equal(t, updatedTheme.Colors.Info, iconColor)

	border, iconColor = model.getColors(Alert{Type: AlertType("custom")})
	assert.Equal(t, updatedTheme.Colors.Border, border)
	assert.Equal(t, updatedTheme.Colors.Foreground, iconColor)
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
