package main

import (
	"context"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/arena"
	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

func arenaHandler() error {
	var tc = arena.TimeControl{
		//FixedTime:  1 * time.Second,
		FixedNodes: 1_500_000,
	}

	var gameConcurrency int
	if tc.FixedNodes != 0 {
		gameConcurrency = runtime.NumCPU()
	} else {
		gameConcurrency = runtime.NumCPU() / 2
	}

	return arena.Run(context.Background(), gameConcurrency, tc, newArenaEngine)
}

func newArenaEngine(experiment bool) arena.IEngine {
	var options = engine.NewMainOptions(evalbuilder.Get(""))
	options.Hash = 128
	options.ExperimentSettings = experiment
	var eng = engine.NewEngine(options)
	eng.Prepare()
	return eng
}
