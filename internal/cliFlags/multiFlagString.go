package cliFlags

import (
	"strings"
)

type MultiFlag []string

func (m *MultiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *MultiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}
