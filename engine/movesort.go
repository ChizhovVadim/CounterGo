package engine

import . "github.com/ChizhovVadim/CounterGo/common"

type orderedMove struct {
	move Move
	key  int
}

func initMovesQS(pos *Position, genChecks bool,
	ml []Move, moves []orderedMove, historyTable historyTable) []orderedMove {
	if pos.IsCheck() {
		ml = pos.GenerateMoves(ml)
	} else {
		ml = pos.GenerateCaptures(ml, genChecks)
	}
	for _, m := range ml {
		var score int
		if isCaptureOrPromotion(m) {
			score = 29000 + mvvlva(m)
		} else {
			score = historyTable.Score(pos.WhiteMove, m)
		}
		moves = append(moves, orderedMove{m, score})
	}
	sortMoves(moves)
	return moves
}

type moveSort struct {
	node      *node
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
			var pos = &ms.node.position
			ms.important = ms.node.buffer1[:0]
			ms.remaining = ms.node.buffer2[:0]
			for _, m := range pos.GenerateMoves(ms.node.buffer0[:]) {
				if m == ms.trans {
					ms.important = append(ms.important, orderedMove{m, 30000})
				} else if isCaptureOrPromotion(m) {
					if seeGEZero(pos, m) {
						var score = 29000 + mvvlva(m)
						ms.important = append(ms.important, orderedMove{m, score})
					} else {
						ms.remaining = append(ms.remaining, orderedMove{m, 0})
					}
				} else if m == ms.node.killer1 {
					ms.important = append(ms.important, orderedMove{m, 28000})
				} else if m == ms.node.killer2 {
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
			var side = ms.node.position.WhiteMove
			var ht = ms.node.engine.historyTable
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

func initMoves(pos *Position, ml []Move, moves []orderedMove,
	trans, killer1, killer2 Move, historyTable historyTable) []orderedMove {
	for _, m := range pos.GenerateMoves(ml) {
		var score int
		if m == trans {
			score = 30000
		} else if isCaptureOrPromotion(m) {
			if seeGEZero(pos, m) {
				score = 29000 + mvvlva(m)
			} else {
				score = historyTable.Score(pos.WhiteMove, m)
			}
		} else if m == killer1 {
			score = 28000
		} else if m == killer2 {
			score = 28000 - 1
		} else {
			score = historyTable.Score(pos.WhiteMove, m)
		}
		moves = append(moves, orderedMove{m, score})
	}
	sortMoves(moves)
	return moves
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
