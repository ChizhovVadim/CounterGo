package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

var logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

func main() {
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	var (
		tacticTestsFilepath = mapPath("~/chess/tests/tests.epd")
		validationPath      = mapPath("~/chess/tuner/quiet-labeled.epd")
		evalName            = "weiss"
	)

	var cli = NewCli()
	cli.AddCommand("benchmark", func() error {
		var path = cli.Params().GetString("testpath", tacticTestsFilepath)
		var evalName = cli.Params().GetString("eval", evalName)
		return runBenchmark(path, evalName)
	})
	cli.AddCommand("tactic", func() error {
		var path = cli.Params().GetString("testpath", tacticTestsFilepath)
		var evalName = cli.Params().GetString("eval", evalName)
		var moveTime = cli.Params().GetInt("movetime", 3)
		return runSolveTactic(path, evalName, time.Duration(moveTime)*time.Second)
	})
	cli.AddCommand("quality", func() error {
		var evalName = cli.Params().GetString("eval", evalName)
		var validationPath = cli.Params().GetString("vd", validationPath)
		return runCheckEvalQuality(evalName, validationPath)
	})
	cli.AddCommand("see", func() error {
		var validationPath = cli.Params().GetString("vd", validationPath)
		return runTestSee(validationPath)
	})
	cli.AddCommand("perft", func() error {
		return runPerft()
	})
	cli.AddCommand("profile", func() error {
		var evalName = cli.Params().GetString("eval", evalName)
		return runProfile(mapPath("~/cpu.prof"), evalName)
	})
	return cli.Execute()
}

type Evaluator interface {
	Evaluate(p *common.Position) int
}

type UciEngine interface {
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

func newEngine(evalName string) *engine.Engine {
	var options = engine.NewMainOptions(evalbuilder.Get(evalName))
	options.Hash = 128
	var eng = engine.NewEngine(options)
	return eng
}
