package stdout

import (
	"fmt"
)

func Warnings(warnings []string) {
	for _, warn := range warnings {
		fmt.Printf("%s %s", YellowS("Warning:"), warn)
	}
}

func Warning(warning string) {
	fmt.Printf("%s %s", YellowS("Warning:"), warning)
}

func Error(msg string) {
	fmt.Printf("%s %s", RedB("Error:"), msg)
}
