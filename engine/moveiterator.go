package engine

const (
	StageImportant = iota
	StageRemaining
)

type moveIterator struct {
	buffer      [MAX_MOVES]Move
	important   []moveWithScore
	remaining   []moveWithScore
	stage, head int
}

type moveWithScore struct {
	Move  Move
	Score int
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
	for _, m := range GenerateCaptures(ctx.Position, genChecks, ctx.mi.buffer[:]) {
		if IsCaptureOrPromotion(m) {
			ctx.mi.important = append(ctx.mi.important, moveWithScore{m, 29000 + MVVLVA(m)})
		} else {
			ctx.mi.remaining = append(ctx.mi.remaining, moveWithScore{m, 0})
		}
	}
	sortMoves(ctx.mi.important)
}

func (ctx *searchContext) InitMoves(hashMove Move) {
	ctx.mi.important = ctx.mi.important[:0]
	ctx.mi.remaining = ctx.mi.remaining[:0]
	ctx.mi.stage = StageImportant
	ctx.mi.head = 0
	for _, m := range GenerateMoves(ctx.Position, ctx.mi.buffer[:]) {
		if m == hashMove {
			ctx.mi.important = append(ctx.mi.important, moveWithScore{m, 30000})
		} else if IsCaptureOrPromotion(m) {
			ctx.mi.important = append(ctx.mi.important, moveWithScore{m, 29000 + MVVLVA(m)})
		} else if m == ctx.Killer1 {
			ctx.mi.important = append(ctx.mi.important, moveWithScore{m, 28000})
		} else if m == ctx.Killer2 {
			ctx.mi.important = append(ctx.mi.important, moveWithScore{m, 28000 - 1})
		} else {
			ctx.mi.remaining = append(ctx.mi.remaining, moveWithScore{m, 0})
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
			for i := range ctx.mi.remaining {
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

func sortMoves(moves []moveWithScore) {
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

func isSorted(moves []moveWithScore) bool {
	for i := 1; i < len(moves); i++ {
		if moves[i-1].Score < moves[i].Score {
			return false
		}
	}
	return true
}
