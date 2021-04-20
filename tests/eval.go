package main

import (
	"fmt"

	"github.com/ChizhovVadim/CounterGo/common"
)

type Evaluator interface {
	Evaluate(p *common.Position) int
}

func testSymmetricEval(e Evaluator) testFenFunc {
	return func(fen string) error {
		var p1, err = common.NewPositionFromFEN(fen)
		if err != nil {
			panic(err)
		}
		var score1 = e.Evaluate(&p1)
		var p2 = common.MirrorPosition(&p1)
		var score2 = e.Evaluate(&p2)
		if score1 != score2 {
			return fmt.Errorf("%v %v", score1, score2)
		}
		return nil
	}
}
