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
	return cli.Execute()
}

type Evaluator interface {
	Evaluate(p *common.Position) int
}

type UciEngine interface {
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

func newEngine(evalName string) *engine.Engine {
	var eng = engine.NewEngine(evalbuilder.Get(evalName))
	eng.Hash = 128
	eng.ExperimentSettings = false
	return eng
}
