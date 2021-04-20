package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/ChizhovVadim/CounterGo/engine"
	"github.com/ChizhovVadim/CounterGo/eval"
	evalpesto "github.com/ChizhovVadim/CounterGo/evalpesto"
)

//TODO use user folder
const TuneFile = "/home/vadim/chess/tuner/quiet-labeled.epd"
const wacFilePath = "/home/vadim/chess/tests/tests.epd"

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	//runTuner()
	//testFens("TestSymmetricEval", testSymmetricEval(eval.NewEvaluationService()))
	//testFens("TestSymmetricEval", testSymmetricEval(evalpesto.NewEvaluationService()))
	//printErr(runBenchmark(false))
	//printErr(runSolveTactic())
}

func runTuner() {
	var t = &Tuner{
		Logger: logger,
		EvalBuilder: func() TunableEvaluator {
			return eval.NewEvaluationService()
		},
		FilePath: TuneFile,
	}
	printErr(t.Run())
}

//go tool pprof counterwork ~/counter.prof
func runBenchmark(profile bool) error {
	if profile {
		f, err := os.Create("/home/vadim/counter.prof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	var eng = newEngine(false)
	var tests, err = loadEpd(wacFilePath)
	if err != nil {
		return err
	}
	benchmark(tests, eng)
	return nil
}

func runSolveTactic() error {
	var eng = newEngine(false)
	var tests, err = loadEpd(wacFilePath)
	if err != nil {
		return err
	}
	solveTactic(tests, eng, 3*time.Second)
	return nil
}

func newEngine(pesto bool) *engine.Engine {
	var evalBuilder func() engine.Evaluator
	if pesto {
		evalBuilder = func() engine.Evaluator {
			return evalpesto.NewEvaluationService()
		}
	} else {
		evalBuilder = func() engine.Evaluator {
			return eval.NewEvaluationService()
		}
	}
	var e = engine.NewEngine(evalBuilder)
	e.Hash = 128
	e.ExperimentSettings = false
	return e
}

func printErr(e error) {
	if e != nil {
		fmt.Println(e)
	}
}
