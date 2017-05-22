package engine

import (
	"fmt"
	"sync"
)

func ParallelSearch(searchStacks []*SearchStack,
	searchMove func(ss *SearchStack) bool) {
	defer func() {
		var r = recover()
		if r != nil && r != searchTimeout {
			panic(r)
		}
	}()
	var ss = searchStacks[0]
	searchMove(ss)
	var wg sync.WaitGroup
	var degreeOfParallelism = len(searchStacks)
	for i := 1; i < degreeOfParallelism; i++ {
		wg.Add(1)
		go func(ss *SearchStack) {
			defer func() {
				wg.Done()
				var r = recover()
				if r != nil && r != searchTimeout {
					panic(r)
				}
			}()
			for searchMove(ss) {
			}
		}(searchStacks[i])
	}
	defer wg.Wait()
	for searchMove(ss) {
	}
}

func ComputeThinkTime(limits LimitsType, side bool) int {
	if limits.MoveTime != 0 {
		return limits.MoveTime
	}
	if limits.Infinite || limits.Ponder {
		return 0
	}
	if side {
		return ComputeTimePerMove(limits.WhiteTime, limits.WhiteIncrement, limits.MovesToGo)
	} else {
		return ComputeTimePerMove(limits.BlackTime, limits.BlackIncrement, limits.MovesToGo)
	}
}

func ComputeTimePerMove(mainTime, incTime, movesToGo int) int {
	if movesToGo == 0 || movesToGo > 35 {
		movesToGo = 35
	}
	var moveTime = (mainTime + incTime*(movesToGo-1)) / movesToGo
	moveTime = min(moveTime, mainTime)
	moveTime = max(0, moveTime-200)
	return moveTime
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

func (pv *PrincipalVariation) String() string {
	var s string
	for temp := pv; temp != nil; temp = temp.Tail {
		if s != "" {
			s += " "
		}
		s += temp.Move.String()
	}
	return s
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
		ScoreToUci(si.Score), si.Depth, si.Nodes, si.Time, nps, si.MainLine.String())
}

func SendProgressToUci(si SearchInfo) {
	if si.Time >= 500 || si.Depth >= 5 {
		fmt.Println(si.String())
	}
}

func SendResultToUci(si SearchInfo) {
	fmt.Println(si.String())
	if tail := si.MainLine.Tail; tail != nil && tail.Move != MoveEmpty {
		fmt.Printf("bestmove %v ponder %v\n", si.MainLine.Move, tail.Move)
	} else {
		fmt.Printf("bestmove %v\n", si.MainLine.Move)
	}
}

func CreateStack(positions []*Position) *SearchStack {
	var items = make([]SearchStack, len(positions)+MAX_HEIGHT+1)
	var index = len(positions) - 1
	for i := 0; i < len(items); i++ {
		if i > 0 {
			items[i].Previous = &items[i-1]
		}
		if i < len(items)-1 {
			items[i].Next = &items[i+1]
		}
		if i <= index {
			items[i].Position = positions[i]
		} else {
			items[i].Height = i - index
			items[i].Position = &Position{}
			items[i].MoveList = &MoveList{}
			items[i].QuietsSearched = make([]Move, 0, MAX_MOVES)
		}
	}
	items[index].MoveList = &MoveList{}
	return &items[index]
}

func IsDraw(ss *SearchStack) bool {
	var p = ss.Position

	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	if p.Rule50 > 100 {
		return true
	}

	for temp := ss.Previous; temp != nil && temp.Position.Rule50 != 0; temp = temp.Previous {
		if temp.Position.Key == p.Key {
			return true
		}
	}

	return false
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
