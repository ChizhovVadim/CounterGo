package main

import (
	"context"
	"flag"
	"log"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

type Config struct {
	Concurrency int
}

var config Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	flag.IntVar(&config.Concurrency, "concurrency", 4, "Number of threads")
	flag.Parse()

	log.Printf("%+v", config)

	var arena = &arena{
		threads:  config.Concurrency,
		openings: getOpenings(),
	}
	return arena.Run(context.Background())
}

func newEngineA() IEngine {
	var eng = engine.NewEngine(func() engine.Evaluator {
		return evalbuilder.Build("counter").(engine.Evaluator)
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
	eng.ExperimentSettings = false
	eng.Prepare()
	return eng
}
