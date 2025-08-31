package render

import (
	"github.com/fatih/color"
)

const (
	RepoW   = 24
	BranchW = 25
	UncommW = 12
	AheadW  = 6
	BehindW = 6
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
