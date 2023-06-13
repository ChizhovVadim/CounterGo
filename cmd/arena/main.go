package main

import (
	"context"
	"log"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

// program for playing games between chess engines
func main() {
	var tc = timeControl{
		//FixedTime:  1 * time.Second,
		FixedNodes: 1_500_000,
	}

	var gameConcurrency int
	if tc.FixedNodes != 0 {
		gameConcurrency = runtime.NumCPU()
	} else {
		gameConcurrency = runtime.NumCPU() / 2
	}

	var err = run(context.Background(), gameConcurrency, tc)
	if err != nil {
		log.Println(err)
	}
}

func newEngineA() IEngine {
	return newEngine(false)
}

func newEngineB() IEngine {
	return newEngine(true)
}

func newEngine(experiment bool) IEngine {
	var options = engine.Options{
		EvalBuilder:        evalbuilder.Get("weiss"),
		Hash:               128,
		Threads:            1,
		ReverseFutility:    true,
		Lmp:                true,
		See:                true,
		ExperimentSettings: experiment,
	}
	options.InitLmr(engine.LmrMult)

	var eng = engine.NewEngine(options)
	eng.Prepare()
	return eng
}
