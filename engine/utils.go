package engine

import (
	"math"
	"sync"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	stackSize     = 64
	maxHeight     = stackSize - 1
	valueDraw     = 0
	valueMate     = 30000
	valueInfinity = valueMate + 1
	valueWin      = valueMate - 2*maxHeight
	valueLoss     = -valueWin
)

func parallelDo(degreeOfParallelism int, body func(threadIndex int)) {
	var wg sync.WaitGroup
	for i := 1; i < degreeOfParallelism; i++ {
		wg.Add(1)
		go func(threadIndex int) {
			body(threadIndex)
			wg.Done()
		}(i)
	}
	body(0)
	wg.Wait()
}

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

var (
	pieceValues = [...]int{0, PawnValue, 4 * PawnValue,
		4 * PawnValue, 6 * PawnValue, 12 * PawnValue, 120 * PawnValue}

	pieceValuesSEE = [...]int{0, 1, 4, 4, 6, 12, 120}
)

func moveValue(move Move) int {
	var result = pieceValues[move.CapturedPiece()]
	if move.Promotion() != Empty {
		result += pieceValues[move.Promotion()] - pieceValues[Pawn]
	}
	return result
}

func isCaptureOrPromotion(move Move) bool {
	return move.CapturedPiece() != Empty ||
		move.Promotion() != Empty
}

func isPawnAdvance(move Move, side bool) bool {
	if move.MovingPiece() != Pawn {
		return false
	}
	var rank = Rank(move.To())
	if side {
		return rank >= Rank5
	} else {
		return rank <= Rank4
	}
}

func isDangerCapture(p *Position, m Move) bool {
	if m.CapturedPiece() == Pawn {
		var pawns = p.Pawns & p.PiecesByColor(!p.WhiteMove)
		if (pawns & (pawns - 1)) == 0 {
			return true
		}
	}
	return false
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

func getAttacks(p *Position, to int, side bool, occ uint64) uint64 {
	var att = (PawnAttacks(to, !side) & p.Pawns) |
		(KnightAttacks[to] & p.Knights) |
		(KingAttacks[to] & p.Kings) |
		(BishopAttacks(to, occ) & (p.Bishops | p.Queens)) |
		(RookAttacks(to, occ) & (p.Rooks | p.Queens))
	return p.PiecesByColor(side) & att
}

func getLeastValuableAttacker(p *Position, to int, side bool, occ uint64) (attacker, from int) {
	attacker = Empty
	from = SquareNone
	var att = getAttacks(p, to, side, occ) & occ
	if att == 0 {
		return
	}
	var newTarget = pieceValuesSEE[King] + 1
	for ; att != 0; att &= att - 1 {
		var f = FirstOne(att)
		var piece = p.WhatPiece(f)
		if pieceValuesSEE[piece] < newTarget {
			attacker = piece
			from = f
			newTarget = pieceValuesSEE[piece]
		}
	}
	return
}

func seeGEZero(p *Position, move Move) bool {
	return seeGE(p, move, 0)
}

func seeGE(p *Position, move Move, bound int) bool {
	var piece = move.MovingPiece()
	var score0 = pieceValuesSEE[move.CapturedPiece()]
	if promotion := move.Promotion(); promotion != Empty {
		piece = move.Promotion()
		score0 += pieceValuesSEE[promotion] - pieceValuesSEE[Pawn]
	}
	var to = move.To()
	var occ = p.White ^ p.Black ^ SquareMask[move.From()]
	occ |= SquareMask[to]
	var side = !p.WhiteMove
	var relativeStm = true
	var balance = score0 - bound
	if balance < 0 {
		return false
	}
	balance -= pieceValuesSEE[piece]
	if balance >= 0 {
		return true
	}
	for {
		var nextVictim, from = getLeastValuableAttacker(p, to, side, occ)
		if nextVictim == Empty {
			return relativeStm
		}
		if piece == King {
			return !relativeStm
		}
		occ ^= SquareMask[from]
		piece = nextVictim
		if relativeStm {
			balance += pieceValuesSEE[nextVictim]
		} else {
			balance -= pieceValuesSEE[nextVictim]
		}
		relativeStm = !relativeStm
		if relativeStm == (balance >= 0) {
			return relativeStm
		}
		side = !side
	}
}

func see(pos *Position, mv Move) int {
	var from = mv.From()
	var to = mv.To()
	var pc = mv.MovingPiece()
	var sd = pos.WhiteMove
	var sc = 0
	if mv.CapturedPiece() != Empty {
		sc += pieceValuesSEE[mv.CapturedPiece()]
	}
	if mv.Promotion() != Empty {
		pc = mv.Promotion()
		sc += pieceValuesSEE[pc] - pieceValuesSEE[Pawn]
	}
	var pieces = (pos.White | pos.Black) &^ SquareMask[from]
	sc -= seeRec(pos, !sd, to, pieces, pc)
	return sc
}

func seeRec(pos *Position, sd bool, to int, pieces uint64, cp int) int {
	var bs = 0
	var pc, from = getLeastValuableAttacker(pos, to, sd, pieces)
	if from != SquareNone {
		var sc = pieceValuesSEE[cp]
		if cp != King {
			sc -= seeRec(pos, !sd, to, pieces&^SquareMask[from], pc)
		}
		if sc > bs {
			bs = sc
		}
	}
	return bs
}

func initLmrCrafty() func(d, m int) int {
	const (
		LMR_rdepth = 1   /* leave 1 full ply after reductions    */
		LMR_min    = 1   /* minimum reduction 1 ply              */
		LMR_max    = 15  /* maximum reduction 15 plies           */
		LMR_db     = 1.8 /* depth is 1.8x as important as        */
		LMR_mb     = 1.0 /* moves searched in the formula.       */
		LMR_s      = 2.0 /* smaller numbers increase reductions. */
	)
	var LMR [32][64]int
	for d := 3; d < 32; d++ {
		for m := 2; m < 64; m++ {
			var r = math.Log(float64(d)*LMR_db) * math.Log(float64(m)*LMR_mb) / LMR_s
			LMR[d][m] = Max(Min(int(r), LMR_max), LMR_min)
			LMR[d][m] = Min(LMR[d][m], Max(d-1-LMR_rdepth, 0))
		}
	}
	return func(d, m int) int {
		return LMR[Min(d, 31)][Min(m, 63)]
	}
}
