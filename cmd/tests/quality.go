package main

import (
	"flag"
	"log"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/internal/quality"
)

func qualityHandler(args []string) error {
	var (
		evalName       = ""
		valDatasetPath = mapPath("~/chess/tuner/quiet-labeled.epd")
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&evalName, "eval", evalName, "")
	//TODO rest flags
	flagset.Parse(args)

	log.Println("checkEvalQuality started",
		"evalName", evalName)
	defer log.Println("checkEvalQuality finished")

	var eval = evalbuilder.Get(evalName)().(quality.IEvaluator)
	return quality.RunQuality(eval, valDatasetPath)
}
