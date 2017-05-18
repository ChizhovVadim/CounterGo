package main

import (
	"counter/engine"
	"counter/shell"
	"runtime"
)

func ResolveSearchService(name string) shell.SearchService {
	return &engine.SearchService{
		MoveOrderService:      engine.NewMoveOrderService(),
		TTable:                engine.NewTranspositionTable(4),
		Evaluate:              engine.Evaluate,
		DegreeOfParallelism:   runtime.NumCPU(),
		UseExperimentSettings: false,
	}
}

func main() {
	var uci = shell.NewUciProtocol(ResolveSearchService)
	uci.Run()
}
