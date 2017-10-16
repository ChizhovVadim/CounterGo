package engine

import (
	"bytes"
	"fmt"
	"sync"
)

func ParallelDo(degreeOfParallelism int, body func(threadIndex int)) {
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

func MateIn(height int) int {
	return VALUE_MATE - height
}

func MatedIn(height int) int {
	return -VALUE_MATE + height
}

func ValueToTT(v, height int) int {
	if v >= VALUE_MATE_IN_MAX_HEIGHT {
		return v + height
	}

	if v <= VALUE_MATED_IN_MAX_HEIGHT {
		return v - height
	}

	return v
}

func ValueFromTT(v, height int) int {
	if v >= VALUE_MATE_IN_MAX_HEIGHT {
		return v - height
	}

	if v <= VALUE_MATED_IN_MAX_HEIGHT {
		return v + height
	}

	return v
}

func PVToUci(pv []Move) string {
	var sb bytes.Buffer
	for i, move := range pv {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(move.String())
	}
	return sb.String()
}

func ScoreToUci(v int) string {
	if VALUE_MATED_IN_MAX_HEIGHT < v && v < VALUE_MATE_IN_MAX_HEIGHT {
		return fmt.Sprintf("cp %v", v)
	} else {
		var mate int
		if v > 0 {
			mate = (VALUE_MATE - v + 1) / 2
		} else {
			mate = (-VALUE_MATE - v) / 2
		}
		return fmt.Sprintf("mate %v", mate)
	}
}

func (si *SearchInfo) String() string {
	var nps = si.Nodes * 1000 / (si.Time + 1)
	return fmt.Sprintf("info score %v depth %v nodes %v time %v nps %v pv %v",
		ScoreToUci(si.Score), si.Depth, si.Nodes, si.Time, nps, PVToUci(si.MainLine))
}

func SendProgressToUci(si SearchInfo) {
	if si.Time >= 500 || si.Depth >= 5 {
		fmt.Println(si.String())
	}
}

func SendResultToUci(si SearchInfo) {
	fmt.Println(si.String())
	if len(si.MainLine) > 0 {
		fmt.Printf("bestmove %v\n", si.MainLine[0])
	}
}

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
	var ownPieces = p.piecesByColor(side)
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

func IsDirectCheck(p *Position, move Move) bool {
	var attacks, occ uint64
	var to = move.To()
	switch move.MovingPiece() {
	case Pawn:
		attacks = PawnAttacks(to, p.WhiteMove)
	case Knight:
		attacks = knightAttacks[to]
	case Bishop:
		occ = ((p.White | p.Black) & ^squareMask[move.From()]) | squareMask[to]
		attacks = BishopAttacks(to, occ)
	case Rook:
		occ = ((p.White | p.Black) & ^squareMask[move.From()]) | squareMask[to]
		attacks = RookAttacks(to, occ)
	case Queen:
		occ = ((p.White | p.Black) & ^squareMask[move.From()]) | squareMask[to]
		attacks = QueenAttacks(to, occ)
	default:
		attacks = 0
	}
	return (attacks & p.Kings & p.piecesByColor(!p.WhiteMove)) != 0
}

func IsCaptureOrPromotion(move Move) bool {
	return move.CapturedPiece() != Empty ||
		move.Promotion() != Empty
}

func IsPawnPush(move Move, side bool) bool {
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

func IsActiveMove(p *Position, move Move) bool {
	if IsCaptureOrPromotion(move) {
		return true
	}
	if IsPawnPush(move, p.WhiteMove) {
		return true
	}
	if IsPassedPawnMove(p, move) {
		return true
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

func HasPawnOn7th(p *Position, side bool) bool {
	if side {
		return (p.Pawns & p.White & Rank7Mask) != 0
	}
	return (p.Pawns & p.Black & Rank2Mask) != 0
}

func IsPassedPawnMove(p *Position, move Move) bool {
	if move.MovingPiece() != Pawn {
		return false
	}

	var file = File(move.To())
	var rank = Rank(move.To())

	if p.WhiteMove {
		if (thisAndNeighboringFiles[file] & upperRanks[rank] & p.Pawns & p.Black) == 0 {
			return true
		}
	} else {
		if (thisAndNeighboringFiles[file] & lowerRanks[rank] & p.Pawns & p.White) == 0 {
			return true
		}
	}

	return false
}

func GetAttacks(p *Position, to int, side bool, occ uint64) uint64 {
	var att = (PawnAttacks(to, !side) & p.Pawns) |
		(knightAttacks[to] & p.Knights) |
		(kingAttacks[to] & p.Kings) |
		(BishopAttacks(to, occ) & (p.Bishops | p.Queens)) |
		(RookAttacks(to, occ) & (p.Rooks | p.Queens))
	return p.piecesByColor(side) & att
}

func SEE_Exchange(p *Position, to int, side bool,
	currScore int, target int, occ uint64) int {
	var att = GetAttacks(p, to, side, occ) & occ
	if att == 0 {
		return currScore
	}

	var from = SquareNone
	var newTarget = pieceValuesSEE[King] + 1

	for ; att != 0; att &= att - 1 {
		var f = FirstOne(att)
		var piece = p.WhatPiece(f)
		if pieceValuesSEE[piece] < newTarget {
			from = f
			newTarget = pieceValuesSEE[piece]
		}
	}

	occ ^= squareMask[from]
	var score = -SEE_Exchange(p, to, !side, -(currScore + target), newTarget, occ)
	if score > currScore {
		return score
	}
	return currScore
}

func SEE(p *Position, move Move) int {
	var from = move.From()
	var to = move.To()
	var piece = move.MovingPiece()
	var captured = move.CapturedPiece()
	var promotion = move.Promotion()
	var side = p.WhiteMove

	var score0 = pieceValuesSEE[captured]
	if promotion != Empty {
		score0 += pieceValuesSEE[promotion] - pieceValuesSEE[Pawn]
		piece = promotion
	}

	var occ = (p.White | p.Black) ^ squareMask[from]
	var score = -SEE_Exchange(p, to, !side, -score0, pieceValuesSEE[piece], occ)
	return score
}
