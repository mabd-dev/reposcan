package alerts

// AlertType each type will be render differently on ui
type AlertType string

var (
	AlertTypeInfo    AlertType = "info"
	AlertTypeWarning AlertType = "warning"
	MsgTypeError     AlertType = "error"
)

// Alert hold data, that will appear later on ui
type Alert struct {
	Type    AlertType
	Title   string
	Message string
	Visible bool
	ticks   int
}

type AlertState struct {
	AlertView string
	X         int
	Y         int
	IsVisible bool
}

type TickMsg struct{}

type AddAlertMsg struct {
	Msg Alert
}
