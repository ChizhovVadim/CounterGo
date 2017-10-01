package main

import (
	"github.com/ChizhovVadim/CounterGo/shell"
)

func main() {
	var uci = shell.NewUciProtocol(shell.NewCounterEngine())
	uci.Run()
}
