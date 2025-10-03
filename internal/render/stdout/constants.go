package stdout

import (
	"github.com/fatih/color"
)

const (
	RepoW        = 24
	BranchW      = 30
	UncommW      = 3
	AheadW       = 3
	BehindW      = 3
	RemoteStateW = 3 + 3 + 3 + 4 //(uncommited files count + aheadW + behindW + 4 space)
)

var (
	BoldS    = color.New(color.Bold).SprintfFunc()
	DimS     = color.New(color.Faint).SprintfFunc()
	GrayS    = color.New(color.FgHiBlack).SprintfFunc()
	CyanBold = color.New(color.FgCyan, color.Bold).SprintfFunc()
	MagBold  = color.New(color.FgMagenta, color.Bold).SprintfFunc()
	BlueS    = color.New(color.FgBlue).SprintfFunc()
	RedS     = color.New(color.FgRed).SprintfFunc()
	RedB     = color.New(color.FgRed, color.Bold).SprintfFunc()
	GreenS   = color.New(color.FgGreen).SprintfFunc()
	YellowS  = color.New(color.FgYellow).SprintfFunc()
)
