package engine

import (
	. "github.com/ChizhovVadim/CounterGo/common"
)

func (ctx *searchContext) Next() *searchContext {
	return &ctx.Engine.tree[ctx.Thread][ctx.Height+1]
}

func (ctx *searchContext) NextOnThread(thread int) *searchContext {
	return &ctx.Engine.tree[thread][ctx.Height+1]
}

func (ctx *searchContext) IsDraw() bool {
	var p = ctx.Position

	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	if p.Rule50 > 100 {
		return true
	}

	var stacks = ctx.Engine.tree[ctx.Thread]
	for i := ctx.Height - 1; i >= 0; i-- {
		var temp = stacks[i].Position
		if temp.Key == p.Key {
			return true
		}
		if temp.Rule50 == 0 || temp.LastMove == MoveEmpty {
			return false
		}
	}

	for _, key := range ctx.Engine.historyKeys {
		if key == p.Key {
			return true
		}
	}

	return false
}

func (ctx *searchContext) ClearPV() {
	ctx.PrincipalVariation = ctx.PrincipalVariation[:0]
}

func (ctx *searchContext) BestMove() Move {
	if len(ctx.PrincipalVariation) == 0 {
		return MoveEmpty
	}
	return ctx.PrincipalVariation[0]
}

func (ctx *searchContext) ComposePV(move Move, child *searchContext) {
	ctx.PrincipalVariation = append(append(ctx.PrincipalVariation[:0], move), child.PrincipalVariation...)
}

func IsLateEndgame(p *Position, side bool) bool {
	//sample: position fen 8/8/6p1/1p2pk1p/1Pp1p2P/2PbP1P1/3N1P2/4K3 w - - 12 58
	var ownPieces = p.PiecesByColor(side)
	return ((p.Rooks|p.Queens)&ownPieces) == 0 &&
		!MoreThanOne((p.Knights|p.Bishops)&ownPieces)
}

var (
	pieceValues = [...]int{0, PawnValue, KnightValue,
		BishopValue, RookValue, QueenValue, QueenValue * 10}

	pieceValuesSEE = [...]int{0, 1, 4, 4, 6, 12, 120}
)

func MoveValue(move Move) int {
	var result = pieceValues[move.CapturedPiece()]
	if move.Promotion() != Empty {
		result += pieceValues[move.Promotion()] - pieceValues[Pawn]
	}
	return result
}

func IsCaptureOrPromotion(move Move) bool {
	return move.CapturedPiece() != Empty ||
		move.Promotion() != Empty
}

func IsPawnAdvance(move Move, side bool) bool {
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

func IsDangerCapture(p *Position, m Move) bool {
	if m.CapturedPiece() == Pawn {
		var pawns = p.Pawns & p.PiecesByColor(!p.WhiteMove)
		if (pawns & (pawns - 1)) == 0 {
			return true
		}
	}
	return false
}

func IsPawnPush7th(move Move, side bool) bool {
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

func GetAttacks(p *Position, to int, side bool, occ uint64) uint64 {
	var att = (PawnAttacks(to, !side) & p.Pawns) |
		(KnightAttacks[to] & p.Knights) |
		(KingAttacks[to] & p.Kings) |
		(BishopAttacks(to, occ) & (p.Bishops | p.Queens)) |
		(RookAttacks(to, occ) & (p.Rooks | p.Queens))
	return p.PiecesByColor(side) & att
}

func GetLeastValuableAttacker(p *Position, to int, side bool, occ uint64) (attacker, from int) {
	attacker = Empty
	from = SquareNone
	var att = GetAttacks(p, to, side, occ) & occ
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

func SEE_GE(p *Position, move Move) bool {
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
	var balance = score0 - pieceValuesSEE[piece]
	if balance >= 0 {
		return true
	}
	for {
		var nextVictim, from = GetLeastValuableAttacker(p, to, side, occ)
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
