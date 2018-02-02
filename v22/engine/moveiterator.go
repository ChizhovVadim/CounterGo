package engine

import . "github.com/ChizhovVadim/CounterGo/common"

const (
	StageImportant = iota
	StageRemaining
)

type moveIterator struct {
	important   []MoveWithScore
	remaining   []MoveWithScore
	stage, head int
}

func (ctx *searchContext) InitQMoves(genChecks bool) {
	if ctx.Position.IsCheck() {
		ctx.InitMoves(MoveEmpty)
		return
	}
	ctx.mi.important = ctx.mi.important[:0]
	ctx.mi.remaining = ctx.mi.remaining[:0]
	ctx.mi.stage = StageImportant
	ctx.mi.head = 0
	ctx.MoveList.GenerateCaptures(ctx.Position, genChecks)
	for i := 0; i < ctx.MoveList.Count; i++ {
		var m = ctx.MoveList.Items[i].Move
		if IsCaptureOrPromotion(m) {
			ctx.mi.important = append(ctx.mi.important, MoveWithScore{m, 29000 + MVVLVA(m)})
		} else {
			ctx.mi.remaining = append(ctx.mi.remaining, MoveWithScore{m, 0})
		}
	}
	sortMoves(ctx.mi.important)
}

func (ctx *searchContext) InitMoves(hashMove Move) {
	ctx.mi.important = ctx.mi.important[:0]
	ctx.mi.remaining = ctx.mi.remaining[:0]
	ctx.mi.stage = StageImportant
	ctx.mi.head = 0
	ctx.MoveList.GenerateMoves(ctx.Position)
	for i := 0; i < ctx.MoveList.Count; i++ {
		var m = ctx.MoveList.Items[i].Move
		if m == hashMove {
			ctx.mi.important = append(ctx.mi.important, MoveWithScore{m, 30000})
		} else if IsCaptureOrPromotion(m) {
			if SEE_GE(ctx.Position, m) {
				ctx.mi.important = append(ctx.mi.important, MoveWithScore{m, 29000 + MVVLVA(m)})
			} else {
				ctx.mi.remaining = append(ctx.mi.remaining, MoveWithScore{m, 0})
			}
		} else if m == ctx.Killer1 {
			ctx.mi.important = append(ctx.mi.important, MoveWithScore{m, 28000})
		} else if m == ctx.Killer2 {
			ctx.mi.important = append(ctx.mi.important, MoveWithScore{m, 28000 - 1})
		} else {
			ctx.mi.remaining = append(ctx.mi.remaining, MoveWithScore{m, 0})
		}
	}
	sortMoves(ctx.mi.important)
}

func (ctx *searchContext) NextMove() Move {
	for {
		switch ctx.mi.stage {
		case StageImportant:
			if ctx.mi.head < len(ctx.mi.important) {
				var m = ctx.mi.important[ctx.mi.head].Move
				ctx.mi.head++
				return m
			}
			ctx.mi.head = 0
			ctx.mi.stage++
			for i := 0; i < len(ctx.mi.remaining); i++ {
				var item = &ctx.mi.remaining[i]
				item.Score = ctx.Engine.historyTable.Score(ctx.Position.WhiteMove, item.Move)
			}
			sortMoves(ctx.mi.remaining)
		case StageRemaining:
			if ctx.mi.head < len(ctx.mi.remaining) {
				var m = ctx.mi.remaining[ctx.mi.head].Move
				ctx.mi.head++
				return m
			}
			return MoveEmpty
		}
	}
}

func MVVLVA(move Move) int {
	var captureScore = pieceValuesSEE[move.CapturedPiece()]
	if move.Promotion() != Empty {
		captureScore += pieceValuesSEE[move.Promotion()] - pieceValuesSEE[Pawn]
	}
	return captureScore*8 - move.MovingPiece()
}

var shellSortGaps = [...]int{10, 4, 1}

func sortMoves(moves []MoveWithScore) {
	for _, gap := range shellSortGaps {
		for i := gap; i < len(moves); i++ {
			j, t := i, moves[i]
			for ; j >= gap && moves[j-gap].Score < t.Score; j -= gap {
				moves[j] = moves[j-gap]
			}
			moves[j] = t
		}
	}
}

func isSorted(moves []MoveWithScore) bool {
	for i := 1; i < len(moves); i++ {
		if moves[i-1].Score < moves[i].Score {
			return false
		}
	}
	return true
}
