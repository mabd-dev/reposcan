package cliFlags

import (
	"fmt"
	"strconv"
)

type BoolFlag struct {
	IsSet bool
	Value bool
}

func (b *BoolFlag) Set(val string) error {
	// flag package parses boolean values from strings
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	b.Value = parsed
	b.IsSet = true
	return nil
}

func (b *BoolFlag) String() string {
	if !b.IsSet {
		return "<unset>"
	}
	return fmt.Sprintf("%v", b.Value)
}

func (b *BoolFlag) IsBoolFlag() bool {
	// This tells the flag package that `-flag` alone means true
	return true
}
