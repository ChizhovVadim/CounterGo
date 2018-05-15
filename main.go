package main

import (
	"github.com/ChizhovVadim/CounterGo/engine"
	"github.com/ChizhovVadim/CounterGo/shell"
)

func main() {
	var engineService = engine.NewEngine()
	var evaluationService = engine.NewEvaluationService()
	evaluationService.TraceEnabled = true
	var uci = shell.NewUciProtocol(engineService, evaluationService)
	uci.Run()
}
