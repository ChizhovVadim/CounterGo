package engine

import (
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
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

func newUciScore(v int) UciScore {
	if v >= valueWin {
		return UciScore{Mate: (valueMate - v + 1) / 2}
	} else if v <= valueLoss {
		return UciScore{Mate: (-valueMate - v) / 2}
	} else {
		return UciScore{Centipawns: v}
	}
}

func isLateEndgame(p *Position, side bool) bool {
	//sample: position fen 8/8/6p1/1p2pk1p/1Pp1p2P/2PbP1P1/3N1P2/4K3 w - - 12 58
	var ownPieces = p.PiecesByColor(side)
	return ((p.Rooks|p.Queens)&ownPieces) == 0 &&
		!MoreThanOne((p.Knights|p.Bishops)&ownPieces)
}

func isCaptureOrPromotion(move Move) bool {
	return move.CapturedPiece() != Empty ||
		move.Promotion() != Empty
}

func isPawnPush7th(move Move, side bool) bool {
	if move.MovingPiece() != Pawn {
		return false
	}
	var rank = Rank(move.To())
	if side {
		return rank == Rank7
	} else {
		return rank == Rank2
	}
}

func isPawnAdvance(move Move, side bool) bool {
	if move.MovingPiece() != Pawn {
		return false
	}
	var rank = Rank(move.To())
	if side {
		return rank >= Rank6
	} else {
		return rank <= Rank3
	}
}

func isRecapture(prev, move Move) bool {
	return prev != MoveEmpty && isCaptureOrPromotion(prev) && move.To() == prev.To()
}
