package alerts

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *AlertModel) renderAlert(alert Alert) string {
	borderColor, iconColor := m.getColors(alert)
	icon := m.getIcon(alert)

	var sb strings.Builder
	if title := m.renderTitle(alert, icon, iconColor); title != "" {
		sb.WriteString(title)
	}

	if alert.Title != "" && alert.Message != "" {
		sb.WriteString("\n")
	}

	if msg := m.renderMessage(alert); msg != "" {
		sb.WriteString(msg)
	}

	// Create styles for this alert
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(sb.String())
}

func (m *AlertModel) getColors(alert Alert) (borderColor lipgloss.Color, iconColor lipgloss.Color) {
	switch alert.Type {
	case MsgTypeError:
		borderColor = m.theme.Colors.Error
		iconColor = m.theme.Colors.Error
	case AlertTypeWarning:
		borderColor = m.theme.Colors.Warning
		iconColor = m.theme.Colors.Warning
	case AlertTypeInfo:
		borderColor = m.theme.Colors.Info
		iconColor = m.theme.Colors.Info
	default:
		borderColor = m.theme.Colors.Border
		iconColor = m.theme.Colors.Foreground
	}
	return borderColor, iconColor
}

func (m *AlertModel) getIcon(tm Alert) (icon string) {
	switch tm.Type {
	case MsgTypeError:
		return "✗"
	case AlertTypeWarning:
		return "⚠"
	case AlertTypeInfo:
		return "ℹ"
	default:
		return "•"
	}
}

func (m *AlertModel) renderTitle(
	alert Alert,
	icon string,
	iconColor lipgloss.Color,
) string {
	if alert.Title == "" {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(iconColor).
		Bold(true)

	return titleStyle.Render(icon + " " + alert.Title)

}

func (m *AlertModel) renderMessage(alert Alert) string {
	if alert.Message == "" {
		return ""
	}

	var messageStyle lipgloss.Style
	if alert.Title != "" {
		messageStyle = lipgloss.NewStyle().
			Foreground(m.theme.Colors.Foreground).
			MarginLeft(2)
	} else {
		messageStyle = lipgloss.NewStyle().
			Foreground(m.theme.Colors.Foreground)
	}

	// timePassed := int(time.Since(alert.addedDate).Seconds())
	// message := messageStyle.Render(alert.Message + " - " + strconv.Itoa(alert.ticks/ticksPerSecond))

	message := messageStyle.Render(alert.Message)
	return message
}
