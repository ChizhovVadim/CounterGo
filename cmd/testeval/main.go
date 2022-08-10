package main

import (
	"flag"
	"log"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Evaluator interface {
	Evaluate(p *common.Position) int
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	var eval string
	var validationPath string

	flag.StringVar(&eval, "eval", "", "Eval function to test")
	flag.StringVar(&validationPath, "vd", "", "Path to validation dataset")
	flag.Parse()

	var e = evalbuilder.NewEvalBuilder(eval)().(Evaluator)
	return checkEvalQuality(e, validationPath)
}
