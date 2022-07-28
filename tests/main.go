package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ChizhovVadim/CounterGo/engine"
	eval "github.com/ChizhovVadim/CounterGo/eval/counter"
)

var logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

func main() {
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	curUser, err := user.Current()
	if err != nil {
		return err
	}
	homeDir := curUser.HomeDir
	if homeDir == "" {
		return fmt.Errorf("current user home dir empty")
	}

	var chessDir = filepath.Join(homeDir, "chess")
	var wacFilePath = filepath.Join(chessDir, "tests/tests.epd")

	//return runBenchmark(wacFilePath)
	return runSolveTactic(wacFilePath)
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
	solveTactic(tests, eng, 3*time.Second)
	return nil
}

func newEngine() *engine.Engine {
	var evalBuilder func() engine.Evaluator
	evalBuilder = func() engine.Evaluator {
		return eval.NewEvaluationService()
	}
	var eng = engine.NewEngine(evalBuilder)
	eng.Hash = 128
	eng.ExperimentSettings = false
	return eng
}
