package uci

import (
	"errors"
	"fmt"
	"strconv"
)

type Option interface {
	UciName() string
	UciString() string
	Set(s string) error
}

type BoolOption struct {
	Name  string
	Value *bool
}

func (opt *BoolOption) UciName() string {
	return opt.Name
}

func (opt *BoolOption) UciString() string {
	return fmt.Sprintf("option name %v type %v default %v",
		opt.Name, "check", *opt.Value)
}

func (opt *BoolOption) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*opt.Value = v
	return nil
}

type IntOption struct {
	Name  string
	Min   int
	Max   int
	Value *int
}

func (opt *IntOption) UciName() string {
	return opt.Name
}

func (opt *IntOption) UciString() string {
	return fmt.Sprintf("option name %v type %v default %v min %v max %v",
		opt.Name, "spin", *opt.Value, opt.Min, opt.Max)
}

func (opt *IntOption) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if v < opt.Min || v > opt.Max {
		return errors.New("argument out of range")
	}
	*opt.Value = v
	return nil
}
