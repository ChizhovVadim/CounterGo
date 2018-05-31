package main

import (
	"github.com/ChizhovVadim/CounterGo/engine"
	"github.com/ChizhovVadim/CounterGo/shell"
)

func main() {
	var uci = shell.NewUciProtocol(engine.NewEngine())
	uci.Run()
}
