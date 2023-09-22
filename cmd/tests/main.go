package main

import (
	"context"
	"log"
	"os"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/internal/utils"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

var cliArgs = NewCommandArgs(os.Args)
var tacticTestsPath = mapPath("~/chess/tests/tests.epd")

func main() {
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	var cli = NewCommandHandler()
	cli.Add("tuner", tunerHandler)
	cli.Add("train", trainHandler)
	cli.Add("benchmark", benchmarkHandler)
	cli.Add("tactic", tacticHandler)
	cli.Add("quality", qualityHandler)
	cli.Add("arena", arenaHandler)
	cli.Add("perft", perftHandler)
	cli.Add("play", func() error {
		return utils.PlayCli(newEngine(""))
	})
	return cli.Execute(cliArgs.CommandName())
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
