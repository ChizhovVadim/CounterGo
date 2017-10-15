package engine

func (ss *SearchStack) NextMove() Move {
	if ss.head >= ss.MoveList.Count {
		return MoveEmpty
	}
	var move = ss.MoveList.ElementAt(ss.head)
	ss.head++
	return move
}

func (ss *SearchStack) InitQMoves(genChecks bool) {
	if ss.Position.IsCheck() {
		ss.InitMoves(MoveEmpty)
		return
	}
	ss.MoveList.GenerateCaptures(ss.Position, genChecks)
	ss.head = 0
	var mos = ss.searchService.HistoryTable
	for i := 0; i < ss.MoveList.Count; i++ {
		var item = &ss.MoveList.Items[i]
		var m = item.Move
		var score int
		if IsCaptureOrPromotion(m) {
			item.Score = 29000 + MVVLVA(m)
		} else {
			item.Score = mos.Score(ss.Position.WhiteMove, m)
		}
		item.Score = score
	}
}

func (ss *SearchStack) InitMoves(hashMove Move) {
	ss.MoveList.GenerateMoves(ss.Position)
	ss.head = 0
	var mos = ss.searchService.HistoryTable
	for i := 0; i < ss.MoveList.Count; i++ {
		var item = &ss.MoveList.Items[i]
		var m = item.Move
		if m == hashMove {
			item.Score = 30000
		} else if IsCaptureOrPromotion(m) {
			item.Score = 29000 + MVVLVA(m)
		} else if m == ss.KillerMove {
			item.Score = 28000
		} else {
			item.Score = mos.Score(ss.Position.WhiteMove, m)
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
