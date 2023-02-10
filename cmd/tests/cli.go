package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type CommandParams struct {
	name   string
	params map[string]string
}

func (params *CommandParams) CommandName() string {
	return params.name
}

func (params *CommandParams) GetString(name string, defaultVal string) string {
	var val, ok = params.params[name]
	if !ok {
		return defaultVal
	}
	return val
}

func (params *CommandParams) GetInt(name string, defaultVal int) int {
	var val, ok = params.params[name]
	if !ok {
		return defaultVal
	}
	var v, err = strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return v
}

func (params *CommandParams) GetDate(name string, defaultVal time.Time) time.Time {
	return defaultVal
}

type Cli struct {
	params   *CommandParams
	handlers map[string]func() error
}

func NewCli() *Cli {
	var args = os.Args
	var cmdName = ""
	var flags = make(map[string]string)
	for i := 1; i < len(args); i++ {
		var arg = args[i]
		if strings.HasPrefix(arg, "-") {
			if i < len(args)-1 {
				var k = strings.TrimPrefix(arg, "-")
				var v = args[i+1]
				flags[k] = v
			}
		} else if cmdName == "" {
			cmdName = arg
		}
	}
	var cmdParams = &CommandParams{
		name:   cmdName,
		params: flags,
	}
	return &Cli{
		params:   cmdParams,
		handlers: map[string]func() error{},
	}
}

func (cli *Cli) Params() *CommandParams {
	return cli.params
}

func (cli *Cli) AddCommand(name string, handler func() error) {
	cli.handlers[name] = handler
}

func (cli *Cli) Execute() error {
	var cmdName = cli.Params().CommandName()
	handler, found := cli.handlers[cmdName]
	if !found {
		return fmt.Errorf("command not found %v", cmdName)
	}
	return handler()
}
