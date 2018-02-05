package engine

import . "github.com/ChizhovVadim/CounterGo/common"

type moveSortQS struct {
	ctx       *searchContext
	genChecks bool
	moves     []orderedMove
	state     int
	head      int
}

type orderedMove struct {
	move Move
	key  int
}

func (ms *moveSortQS) Next() Move {
	for {
		switch ms.state {
		case 0:
			var ml = ms.ctx.buffer0[:]
			if ms.ctx.position.IsCheck() {
				ml = GenerateMoves(ml, ms.ctx.position)
			} else {
				ml = GenerateCaptures(ml, ms.ctx.position, ms.genChecks)
			}
			ms.moves = ms.ctx.buffer1[:0]
			for _, m := range ml {
				var score int
				if IsCaptureOrPromotion(m) {
					score = 29000 + mvvlva(m)
				} else {
					score = ms.ctx.engine.historyTable.Score(ms.ctx.position.WhiteMove, m)
				}
				ms.moves = append(ms.moves, orderedMove{m, score})
			}
			sortMoves(ms.moves)
			ms.state++
			ms.head = 0
		case 1:
			if ms.head < len(ms.moves) {
				var m = ms.moves[ms.head].move
				ms.head++
				return m
			}
			return MoveEmpty
		}
	}
}

type moveSort struct {
	ctx       *searchContext
	trans     Move
	important []orderedMove
	remaining []orderedMove
	state     int
	head      int
}

func (ms *moveSort) Next() Move {
	for {
		switch ms.state {
		case 0:
			var pos = ms.ctx.position
			ms.important = ms.ctx.buffer1[:0]
			ms.remaining = ms.ctx.buffer2[:0]
			for _, m := range GenerateMoves(ms.ctx.buffer0[:], pos) {
				if m == ms.trans {
					ms.important = append(ms.important, orderedMove{m, 30000})
				} else if IsCaptureOrPromotion(m) {
					var score = 29000 + mvvlva(m)
					if !SEE_GE(pos, m) {
						score -= 500
					}
					ms.important = append(ms.important, orderedMove{m, score})
				} else if m == ms.ctx.killer1 {
					ms.important = append(ms.important, orderedMove{m, 28000})
				} else if m == ms.ctx.killer2 {
					ms.important = append(ms.important, orderedMove{m, 28000 - 1})
				} else {
					ms.remaining = append(ms.remaining, orderedMove{m, 0})
				}
			}
			sortMoves(ms.important)
			ms.state++
			ms.head = 0
		case 1:
			if ms.head < len(ms.important) {
				var m = ms.important[ms.head].move
				ms.head++
				return m
			}
			var side = ms.ctx.position.WhiteMove
			var ht = ms.ctx.engine.historyTable
			for i := range ms.remaining {
				var item = &ms.remaining[i]
				item.key = ht.Score(side, item.move)
			}
			sortMoves(ms.remaining)
			ms.state++
			ms.head = 0
		case 2:
			if ms.head < len(ms.remaining) {
				var m = ms.remaining[ms.head].move
				ms.head++
				return m
			}
			return MoveEmpty
		}
	}
}

func mvvlva(move Move) int {
	var captureScore = pieceValuesSEE[move.CapturedPiece()]
	if move.Promotion() != Empty {
		captureScore += pieceValuesSEE[move.Promotion()] - pieceValuesSEE[Pawn]
	}
	return captureScore*8 - move.MovingPiece()
}

var shellSortGaps = [...]int{10, 4, 1}

func sortMoves(moves []orderedMove) {
	for _, gap := range shellSortGaps {
		for i := gap; i < len(moves); i++ {
			j, t := i, moves[i]
			for ; j >= gap && moves[j-gap].key < t.key; j -= gap {
				moves[j] = moves[j-gap]
			}
			moves[j] = t
		}
	}
}

func isSorted(moves []orderedMove) bool {
	for i := 1; i < len(moves); i++ {
		if moves[i-1].key < moves[i].key {
			return false
		}
	}
	return true
}
