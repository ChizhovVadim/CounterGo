package engine

import (
	"math"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"github.com/ChizhovVadim/CounterGo/pkg/model"
)

const (
	stackSize     = 128
	maxHeight     = stackSize - 1
	valueDraw     = 0
	valueMate     = 30000
	valueInfinity = valueMate + 1
	valueWin      = valueMate - 2*maxHeight
	valueLoss     = -valueWin
)

func winIn(height int) int {
	return valueMate - height
}

func lossIn(height int) int {
	return -valueMate + height
}

func valueToTT(v, height int) int {
	if v >= valueWin {
		return v + height
	}

	if v <= valueLoss {
		return v - height
	}

	return v
}

func valueFromTT(v, height int) int {
	if v >= valueWin {
		return v - height
	}

	if v <= valueLoss {
		return v + height
	}

	return v
}

func newUciScore(v int) model.UciScore {
	if v >= valueWin {
		return model.UciScore{Mate: (valueMate - v + 1) / 2}
	} else if v <= valueLoss {
		return model.UciScore{Mate: (-valueMate - v) / 2}
	} else {
		return model.UciScore{Centipawns: v}
	}
}

func isLateEndgame(p *common.Position, side bool) bool {
	//sample: position fen 8/8/6p1/1p2pk1p/1Pp1p2P/2PbP1P1/3N1P2/4K3 w - - 12 58
	var ownPieces = p.PiecesByColor(side)
	return ((p.Rooks|p.Queens)&ownPieces) == 0 &&
		!common.MoreThanOne((p.Knights|p.Bishops)&ownPieces)
}

func isCaptureOrPromotion(move common.Move) bool {
	return move.CapturedPiece() != common.Empty ||
		move.Promotion() != common.Empty
}

func LmrMult(d, m float64) float64 {
	return lirp(math.Log(d)*math.Log(m), math.Log(5)*math.Log(22), math.Log(63)*math.Log(63), 3, 8)
}

func lirp(x, x1, x2, y1, y2 float64) float64 {
	return y1 + (y2-y1)*(x-x1)/(x2-x1)
}
