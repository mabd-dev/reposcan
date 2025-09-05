package cliFlags

import (
	"fmt"
	"strconv"
)

type IntFlag struct {
	IsSet bool
	Value int
}

func (b *IntFlag) Set(val string) error {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	b.Value = parsed
	b.IsSet = true
	return nil
}

func (b *IntFlag) String() string {
	if !b.IsSet {
		return "<unset>"
	}
	return fmt.Sprintf("%v", b.Value)
}
