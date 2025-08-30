package render

import (
	"fmt"

	. "github.com/MABD-dev/RepoScan/internal/utils"
)

func Warning(warning string) {
	fmt.Printf("%s %s", YellowS("Warning:"), warning)
}

func Error(msg string) {
	fmt.Printf("%s %s", RedB("Error:"), msg)
}
