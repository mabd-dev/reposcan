package tui

import (
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
)

var reposTableKeybindings = []common.Keybinding{
	{
		Key:         "↑/↓",
		Description: "Navigate up and down (or j/k)",
		ShortDesc:   "Navigate",
	},
	{
		Key:         "<enter>",
		Description: "Open git repository report details",
		ShortDesc:   "Details",
	},
	{
		Key:         "c",
		Description: "Copy repo path to clipboard",
		ShortDesc:   "Copy Path",
	},
	{
		Key:         "/",
		Description: "Filter by repo/branch name",
		ShortDesc:   "Filter",
	},
	{
		Key:         "q",
		Description: "Quit",
		ShortDesc:   "Quit",
	},
}

var reposTableFilterKeybindings = []common.Keybinding{
	{
		Key:         "<enter>",
		Description: "Apply and move cursor to repos table",
		ShortDesc:   "Apply",
	},
	{
		Key:         "<esc>",
		Description: "Hide and cancel filter",
		ShortDesc:   "Cancel",
	},
}

var helpPopupKeybindings = []common.Keybinding{
	{
		Key:         "q/<esc>",
		Description: "Close Popup",
		ShortDesc:   "Close",
	},
}
