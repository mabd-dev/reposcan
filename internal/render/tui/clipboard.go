package tui

import "golang.design/x/clipboard"

var copyTextToClipboard = func(text string) bool {
	if err := clipboard.Init(); err != nil {
		return false
	}

	return clipboard.Write(clipboard.FmtText, []byte(text)) != nil
}
