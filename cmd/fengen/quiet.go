package main

import (
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

func isDraw(p *common.Position) bool {
	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!common.MoreThanOne(p.Knights|p.Bishops) {
		return true
	}
	return false
}

func isNoisyPos(p *common.Position, onlyForSideToMove bool) bool {
	if hasGoodCapture(p) {
		return true
	}
	if !onlyForSideToMove && !p.IsCheck() {
		var child common.Position
		p.MakeNullMove(&child)
		if hasGoodCapture(&child) {
			return true
		}
	}
	return false
}

func hasGoodCapture(p *common.Position) bool {
	var buffer [32]common.OrderedMove
	var ml = p.GenerateCaptures(buffer[:])
	for i := range ml {
		var move = ml[i].Move
		if engine.SeeGE(p, move, 0) {
			return true
		}
	}
	return false
}
