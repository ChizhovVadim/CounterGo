package engine

import . "github.com/ChizhovVadim/CounterGo/pkg/common"

const sortTableKeyImportant = 100000

type moveIteratorQS struct {
	position *Position
	buffer   []OrderedMove
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

func (mi *moveIteratorQS) Next() Move {
	if mi.index >= mi.count {
		return MoveEmpty
	}
	var m = mi.buffer[mi.index].Move
	mi.index++
	return m
}

type moveIterator struct {
	position  *Position
	buffer    []OrderedMove
	history   historyContext
	transMove Move
	killer1   Move
	killer2   Move
	count     int
	index     int
}

func (mi *moveIterator) Init() {
	mi.count = len(mi.position.GenerateMoves(mi.buffer))

	var side = mi.position.WhiteMove
	for i := 0; i < mi.count; i++ {
		var m = mi.buffer[i].Move
		var score int
		if m == mi.transMove {
			score = sortTableKeyImportant + 2000
		} else if isCaptureOrPromotion(m) {
			if seeGEZero(mi.position, m) {
				score = sortTableKeyImportant + 1000 + mvvlva(m)
			} else {
				score = 0 + mvvlva(m)
			}
		} else if m == mi.killer1 {
			score = sortTableKeyImportant + 1
		} else if m == mi.killer2 {
			score = sortTableKeyImportant
		} else {
			// ideally should be inlined. copy/paste?
			score = mi.history.ReadTotal(side, m)
		}
		mi.buffer[i].Key = int32(score)
	}
}

func (mi *moveIterator) Reset() {
	mi.index = 0
}

func (mi *moveIterator) Next() Move {
	if mi.index >= mi.count {
		return MoveEmpty
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

var sortPieceValues = [...]int{Empty: 0, Pawn: 1, Knight: 2, Bishop: 3, Rook: 4, Queen: 5, King: 6}

func mvvlva(move Move) int {
	return 8*(sortPieceValues[move.CapturedPiece()]+
		sortPieceValues[move.Promotion()]) -
		sortPieceValues[move.MovingPiece()]
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

func moveToTop(ml []OrderedMove) {
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

func skipQuiets(ml []OrderedMove, startIndex, endIndex int) int {
	var i = startIndex
	for j := startIndex; j < endIndex; j++ {
		if !isCaptureOrPromotion(ml[j].Move) {
			if i != j {
				ml[i], ml[j] = ml[j], ml[i]
			}
			i++
		}
	}
	return i
}
