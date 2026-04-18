package engine

import "github.com/ChizhovVadim/CounterGo/pkg/common"

const sortTableKeyImportant = 100000

type moveIteratorQS struct {
	position *common.Position
	buffer   []common.OrderedMove
	count    int
	index    int
}

func (mi *moveIteratorQS) Init() {
	if mi.position.IsCheck() {
		mi.count = len(mi.position.GenerateMoves(mi.buffer))
	} else {
		mi.count = len(mi.position.GenerateCaptures(mi.buffer))
	}

	for i := 0; i < mi.count; i++ {
		var m = mi.buffer[i].Move
		var score int
		if isCaptureOrPromotion(m) {
			score = 29000 + mvvlva(m)
		} else {
			score = 0
		}
		mi.buffer[i].Key = int32(score)
	}

	sortMoves(mi.buffer[:mi.count])
}

func (mi *moveIteratorQS) Reset() {
	mi.index = 0
}

func (mi *moveIteratorQS) Next() common.Move {
	if mi.index >= mi.count {
		return common.MoveEmpty
	}
	var m = mi.buffer[mi.index].Move
	mi.index++
	return m
}

type moveIterator struct {
	buffer []common.OrderedMove
	count  int
	index  int
}

func (t *thread) initMoveIterator(height int, transMove common.Move) moveIterator {
	var position = &t.stack[height].position
	var killer1 = t.stack[height].killer1
	var killer2 = t.stack[height].killer2
	var buffer = t.stack[height].moveList[:]
	var count = len(position.GenerateMoves(buffer))
	var history = t.getHistoryContext(height)

	for i := 0; i < count; i++ {
		var m = buffer[i].Move
		var score int
		if m == transMove {
			score = sortTableKeyImportant + 2000
		} else if isCaptureOrPromotion(m) {
			if seeGEZero(position, m) {
				score = sortTableKeyImportant + 1000 + mvvlva(m)
			} else {
				score = 0 + mvvlva(m)
			}
		} else if m == killer1 {
			score = sortTableKeyImportant + 1
		} else if m == killer2 {
			score = sortTableKeyImportant
		} else {
			// ideally should be inlined. copy/paste?
			score = history.ReadTotal(m)
		}
		buffer[i].Key = int32(score)
	}

	return moveIterator{
		buffer: buffer,
		count:  count,
	}
}

func (mi *moveIterator) Reset() {
	mi.index = 0
}

func (mi *moveIterator) Next() common.Move {
	if mi.index >= mi.count {
		return common.MoveEmpty
	}
	const SortMovesIndex = 1
	if mi.index <= SortMovesIndex {
		if mi.index == SortMovesIndex {
			sortMoves(mi.buffer[mi.index:mi.count])
		} else {
			moveToTop(mi.buffer[mi.index:mi.count])
		}
	}
	var m = mi.buffer[mi.index].Move
	mi.index++
	return m
}

var sortPieceValues = [...]int{common.Empty: 0, common.Pawn: 1, common.Knight: 2, common.Bishop: 3, common.Rook: 4, common.Queen: 5, common.King: 6}

func mvvlva(move common.Move) int {
	return 8*(sortPieceValues[move.CapturedPiece()]+
		sortPieceValues[move.Promotion()]) -
		sortPieceValues[move.MovingPiece()]
}

func sortMoves(moves []common.OrderedMove) {
	for i := 1; i < len(moves); i++ {
		j, t := i, moves[i]
		for ; j > 0 && moves[j-1].Key < t.Key; j-- {
			moves[j] = moves[j-1]
		}
		moves[j] = t
	}
}

func moveToTop(ml []common.OrderedMove) {
	var bestIndex = 0
	for i := 1; i < len(ml); i++ {
		if ml[i].Key > ml[bestIndex].Key {
			bestIndex = i
		}
	}
	if bestIndex != 0 {
		ml[0], ml[bestIndex] = ml[bestIndex], ml[0]
	}
}
