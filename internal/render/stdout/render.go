package stdout

import (
	"fmt"
)

// Warnings prints a list of warning messages with a highlighted prefix.
func Warnings(warnings []string) {
	for _, warn := range warnings {
		fmt.Printf("%s %s\n", YellowS("Warning:"), warn)
	}
}

// Warning prints a single warning message with a highlighted prefix.
func Warning(warning string) {
	fmt.Printf("%s %s\n", YellowS("Warning:"), warning)
}

// Error prints a single error message with a highlighted prefix.
func Error(msg string) {
	fmt.Printf("%s %s\n", RedB("Error:"), msg)
}
