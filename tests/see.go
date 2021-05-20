package main

import (
	"fmt"

	"github.com/ChizhovVadim/CounterGo/common"
	"github.com/ChizhovVadim/CounterGo/engine"
)

func testSee() testFenFunc {
	var buffer [common.MaxMoves]common.OrderedMove
	var child common.Position
	return func(fen string) error {
		var p, err = common.NewPositionFromFEN(fen)
		if err != nil {
			panic(err)
		}
		var ml = p.GenerateMoves(buffer[:])
		for i := range ml {
			var move = ml[i].Move
			if !p.MakeMove(move, &child) {
				continue
			}
			var see = engine.See(&p, move)
			if !engine.SeeGE(&p, move, see) ||
				engine.SeeGE(&p, move, see+1) {
				return fmt.Errorf("%v", see)
			}
		}
		return nil
	}
}
