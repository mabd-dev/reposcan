// Package alerts handles showing alerts over the ui with auto-disappear functionality
package alerts

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/theme"
)

// 14 ticks, each quarter a second => 3 seconds and the alert will auto-disappear
const alertAppearDurationInSeconds = 3
const ticksPerSecond int = 4
const maxTicks int = alertAppearDurationInSeconds * ticksPerSecond
const timer time.Duration = time.Duration(1_000/ticksPerSecond) * time.Millisecond

type AlertModel struct {
	alerts []Alert
	theme  theme.Theme
}

func New(t theme.Theme) AlertModel {
	return AlertModel{
		theme: t,
	}
}

func (m AlertModel) Update(msg tea.Msg) (AlertModel, tea.Cmd) {
	switch msg := msg.(type) {
	case AddAlertMsg:
		var cmd tea.Cmd
		if len(m.alerts) > 0 {
			cmd = nil
		} else {
			cmd = startTimer(1 * time.Second)
		}

		alert := msg.Msg
		m.addAlert(alert)
		return m, cmd

	case TickMsg:
		m.onTick()

		if len(m.alerts) == 0 {
			return m, nil
		} else {
			return m, startTimer(250 * time.Millisecond)
		}
	}
	return m, nil
}

// AlertViews render each alert, then return all strings to be placed properly
func (m AlertModel) AlertStates(
	totalWidth int,
	totalHeight int,
) []AlertState {
	alertStates := make([]AlertState, 0, len(m.alerts))

	if len(m.alerts) == 0 {
		return []AlertState{}
	}

	var x, y int

	for i, tm := range m.alerts {
		alertView := m.renderAlert(tm)

		alertHeight := lipgloss.Height(alertView)

		x = totalWidth - lipgloss.Width(alertView)
		isVisible := y+alertHeight <= totalHeight

		alertStates = append(alertStates, AlertState{
			AlertView: alertView,
			X:         x,
			Y:         y,
			IsVisible: isVisible,
		})
		m.alerts[i].Visible = isVisible

		y += alertHeight
	}

	return alertStates
}

func (m *AlertModel) addAlert(alert Alert) {
	alert.ticks = maxTicks
	alert.Visible = true
	m.alerts = append(m.alerts, alert)
}

// onTick decrease first alert tick by 1
func (m *AlertModel) onTick() {
	if len(m.alerts) > 0 {
		m.alerts[0].ticks--

		if m.alerts[0].ticks <= 0 {
			m.alerts = m.alerts[1:]
		}
	}

}

func startTimer(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}
