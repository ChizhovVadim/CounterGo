package main

import (
	"CounterGo/shell"
)

func main() {
	var uci = shell.NewUciProtocol(shell.NewCounterEngine())
	uci.Run()
}
