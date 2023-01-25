package main

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

// program for playing games between chess engines
func main() {
	var gameConcurrency = 4
	var tc = timeControl{
		FixedTime:  1 * time.Second,
		FixedNodes: 0,
	}

	var numCPU = runtime.NumCPU()
	if gameConcurrency > numCPU {
		gameConcurrency = numCPU
	}
	runtime.GOMAXPROCS(gameConcurrency)

	var err = run(context.Background(), gameConcurrency, tc)
	if err != nil {
		log.Println(err)
	}
}

func newEngineA() IEngine {
	var eng = engine.NewEngine(func() engine.Evaluator {
		return evalbuilder.Build("weiss").(engine.Evaluator)
	})
	eng.Hash = 128
	eng.Threads = 1
	eng.ExperimentSettings = false
	eng.Prepare()
	return eng
}

func newEngineB() IEngine {
	var eng = engine.NewEngine(func() engine.Evaluator {
		return evalbuilder.Build("weiss").(engine.Evaluator)
	})
	eng.Hash = 128
	eng.Threads = 1
	eng.ExperimentSettings = true
	eng.Prepare()
	return eng
}
