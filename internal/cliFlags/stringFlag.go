package cliFlags

import (
	"fmt"
)

type StringFlag struct {
	IsSet bool
	Value string
}

func (b *StringFlag) Set(val string) error {
	b.Value = val
	b.IsSet = true
	return nil
}

func (b *StringFlag) String() string {
	if !b.IsSet {
		return "<unset>"
	}
	return fmt.Sprintf("%v", b.Value)
}
