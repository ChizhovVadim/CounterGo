package main

import (
	"flag"
	"log"
	"time"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/internal/tactic"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

func tacticHandler(args []string) error {
	var (
		filepath = mapPath("~/chess/tests/tests.epd")
		evalName = ""
		moveTime = 3 * time.Second
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&evalName, "eval", evalName, "")
	//TODO rest flags
	flagset.Parse(args)

	log.Println("solveTactic started",
		"filepath", filepath,
		"evalName", evalName,
		"moveTime", moveTime)
	defer log.Println("solveTactic finished")

	var tests, err = tactic.LoadEpd(filepath)
	if err != nil {
		return err
	}
	var eng = newEngine(evalName)
	eng.Options.ProgressMinNodes = 0
	eng.Prepare()
	return tactic.SolveTactic(tests, eng, moveTime)
}

func newEngine(evalName string) *engine.Engine {
	var options = engine.NewMainOptions(evalbuilder.Get(evalName))
	options.Hash = 128
	var eng = engine.NewEngine(options)
	return eng
}
