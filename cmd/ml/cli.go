package main

import (
	"fmt"
	"strconv"
	"strings"
)

type CommandArgs struct {
	commandName string
	params      map[string]string
}

func NewCommandArgs(args []string) *CommandArgs {
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
	return &CommandArgs{
		commandName: cmdName,
		params:      flags,
	}
}

func (ca *CommandArgs) CommandName() string {
	return ca.commandName
}

func (ca *CommandArgs) GetString(name string, defaultVal string) string {
	var val, ok = ca.params[name]
	if !ok {
		return defaultVal
	}
	return val
}

func (ca *CommandArgs) GetInt(name string, defaultVal int) int {
	var val, ok = ca.params[name]
	if !ok {
		return defaultVal
	}
	var v, err = strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return v
}

type CommandHandler struct {
	items map[string]func() error
}

func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		items: make(map[string]func() error),
	}
}

func (ch *CommandHandler) Add(name string, handler func() error) {
	ch.items[name] = handler
}

func (ch *CommandHandler) Execute(commandName string) error {
	handler, found := ch.items[commandName]
	if !found {
		return fmt.Errorf("command not found %v", commandName)
	}
	return handler()
}
