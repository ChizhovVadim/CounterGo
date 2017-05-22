package main

import (
	"counter/shell"
)

func main() {
	var uci = shell.NewUciProtocol(shell.NewCounterEngine())
	uci.Run()
}
