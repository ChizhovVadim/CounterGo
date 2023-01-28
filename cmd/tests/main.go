package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

var logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

type Config struct {
	name                string
	moveTime            time.Duration
	eval                string
	tacticTestsFilepath string
	validationPath      string
}

var config Config

func main() {
	flag.StringVar(&config.name, "name", "quality", "quality|benchmark|tactic")
	flag.StringVar(&config.eval, "eval", "", "Eval function")
	flag.StringVar(&config.tacticTestsFilepath, "testpath", "", "File path to tactic tests")
	flag.DurationVar(&config.moveTime, "movetime", 3*time.Second, "Time per one tactic test")
	flag.StringVar(&config.validationPath, "vd", "", "Path to validation dataset")
	flag.Parse()

	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	log.Printf("%+v", config)

	if config.name == "benchmark" {
		return runBenchmark(config.tacticTestsFilepath)
	} else if config.name == "tactic" {
		return runSolveTactic(config.tacticTestsFilepath)
	} else if config.name == "quality" {
		var evaluator = evalbuilder.Get(config.eval)().(Evaluator)
		return checkEvalQuality(evaluator, config.validationPath)
	}

	return nil
}

func runBenchmark(filepath string) error {
	var tests, err = loadEpd(filepath)
	if err != nil {
		return err
	}
	var eng = newEngine()
	benchmark(tests, eng)
	return nil
}

func runSolveTactic(filepath string) error {
	var tests, err = loadEpd(filepath)
	if err != nil {
		return err
	}
	var eng = newEngine()
	eng.ProgressMinNodes = 0
	solveTactic(tests, eng, config.moveTime)
	return nil
}

func newEngine() *engine.Engine {
	var eng = engine.NewEngine(evalbuilder.Get(config.eval))
	eng.Hash = 128
	eng.ExperimentSettings = false
	return eng
}
