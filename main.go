package main

import (
	"github.com/ChizhovVadim/CounterGo/shell"
	"github.com/ChizhovVadim/CounterGo/v21/engine"
)

func main() {
	var uci = shell.NewUciProtocol(engine.NewEngine())
	uci.Run()
}
