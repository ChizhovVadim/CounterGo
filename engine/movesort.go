package engine

import . "github.com/ChizhovVadim/CounterGo/common"

const (
	sortTableKeyImportant = 27000
	historyMax            = 1 << 14
)

type sortTable struct {
	killers [stackSize][2]Move
	history [1024]int
	counter [1024]Move
}

func (st *sortTable) Clear() {
	for i := range st.killers {
		st.killers[i][0] = MoveEmpty
		st.killers[i][1] = MoveEmpty
	}
	for i := range st.history {
		st.history[i] = 0
	}
	for i := range st.counter {
		st.counter[i] = MoveEmpty
	}
}

func (st *sortTable) Update(p *Position, bestMove Move, searched []Move, depth, height int) {
	if st.killers[height][0] != bestMove {
		st.killers[height][1] = st.killers[height][0]
		st.killers[height][0] = bestMove
	}
	if p.LastMove != MoveEmpty {
		st.counter[pieceSquareIndex(!p.WhiteMove, p.LastMove)] = bestMove
	}
	var side = p.WhiteMove
	var bonus = Min(depth*depth, 400)
	for _, m := range searched {
		var index = pieceSquareIndex(side, m)
		if m == bestMove {
			st.history[index] += (historyMax - st.history[index]) * bonus / 512
			break
		} else {
			st.history[index] += (-historyMax - st.history[index]) * bonus / 512
		}
	}
}

func (st *sortTable) Note(p *Position, ml []OrderedMove, trans Move, height int) {
	var side = p.WhiteMove
	var killer1 = st.killers[height][0]
	var killer2 = st.killers[height][1]
	var counter Move
	if p.LastMove != MoveEmpty {
		counter = st.counter[pieceSquareIndex(!p.WhiteMove, p.LastMove)]
	}
	for i := range ml {
		var m = ml[i].Move
		var score int
		if m == trans {
			score = 30000
		} else if isCaptureOrPromotion(m) {
			if seeGEZero(p, m) {
				score = 29000 + mvvlva(m)
			} else {
				score = mvvlva(m)
			}
		} else if m == killer1 {
			score = 28000
		} else if m == killer2 {
			score = 28000 - 1
		} else if m == counter {
			score = 28000 - 2
		} else {
			score = st.history[pieceSquareIndex(side, m)]
		}
		ml[i].Key = score
	}
}

func (st *sortTable) NoteQS(p *Position, ml []OrderedMove) {
	var side = p.WhiteMove
	for i := range ml {
		var m = ml[i].Move
		var score int
		if isCaptureOrPromotion(m) {
			score = 29000 + mvvlva(m)
		} else {
			score = st.history[pieceSquareIndex(side, m)]
		}
		ml[i].Key = score
	}
}

func pieceSquareIndex(side bool, move Move) int {
	var result = (move.MovingPiece() << 6) | move.To()
	if side {
		result |= 1 << 9
	}
	return result
}

func mvvlva(move Move) int {
	var captureScore = pieceValuesSEE[move.CapturedPiece()]
	if move.Promotion() != Empty {
		captureScore += pieceValuesSEE[move.Promotion()] - pieceValuesSEE[Pawn]
	}
	return captureScore*8 - move.MovingPiece()
}

func sortMoves(moves []OrderedMove) {
	for i := 1; i < len(moves); i++ {
		j, t := i, moves[i]
		for ; j > 0 && moves[j-1].Key < t.Key; j-- {
			moves[j] = moves[j-1]
		}
		moves[j] = t
	}
}

func isSorted(moves []OrderedMove) bool {
	for i := 1; i < len(moves); i++ {
		if moves[i-1].Key < moves[i].Key {
			return false
		}
	}
	return true
}
