package engine

func (ctx *searchContext) NextMove() Move {
	if ctx.Head >= ctx.MoveList.Count {
		return MoveEmpty
	}
	var move = ctx.MoveList.ElementAt(ctx.Head)
	ctx.Head++
	return move
}

func (ctx *searchContext) InitQMoves(genChecks bool) {
	if ctx.Position.IsCheck() {
		ctx.InitMoves(MoveEmpty)
		return
	}
	ctx.MoveList.GenerateCaptures(ctx.Position, genChecks)
	ctx.Head = 0
	var mos = ctx.Engine.historyTable
	for i := 0; i < ctx.MoveList.Count; i++ {
		var item = &ctx.MoveList.Items[i]
		var m = item.Move
		var score int
		if IsCaptureOrPromotion(m) {
			item.Score = 29000 + MVVLVA(m)
		} else {
			item.Score = mos.Score(ctx.Position.WhiteMove, m)
		}
		item.Score = score
	}
}

func (ctx *searchContext) InitMoves(hashMove Move) {
	ctx.MoveList.GenerateMoves(ctx.Position)
	ctx.Head = 0
	var mos = ctx.Engine.historyTable
	for i := 0; i < ctx.MoveList.Count; i++ {
		var item = &ctx.MoveList.Items[i]
		var m = item.Move
		if m == hashMove {
			item.Score = 30000
		} else if IsCaptureOrPromotion(m) {
			item.Score = 29000 + MVVLVA(m)
		} else if m == ctx.Killer {
			item.Score = 28000
		} else {
			item.Score = mos.Score(ctx.Position.WhiteMove, m)
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
