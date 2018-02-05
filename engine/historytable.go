package engine

import (
	"sync/atomic"

	. "github.com/ChizhovVadim/CounterGo/common"
)

type historyTable []historyEntry

type historyEntry struct {
	success, try int32
}

func NewHistoryTable() historyTable {
	return make([]historyEntry, 1<<10)
}

func (ht historyTable) Clear() {
	for i := range ht {
		ht[i] = historyEntry{1, 1}
	}
}

func (ht historyTable) Update(ctx *searchContext, bestMove Move, depth int) {
	if ctx.killer1 != bestMove {
		ctx.killer2 = ctx.killer1
		ctx.killer1 = bestMove
	}
	var side = ctx.position.WhiteMove
	for _, move := range ctx.quietsSearched {
		atomic.AddInt32(&ht[pieceSquareIndex(side, move)].try, int32(depth))
	}
	atomic.AddInt32(&ht[pieceSquareIndex(side, bestMove)].success, int32(depth))
}

func (ht historyTable) Score(side bool, move Move) int {
	var entry = ht[pieceSquareIndex(side, move)]
	return int((entry.success << 10) / entry.try)
}

func pieceSquareIndex(side bool, move Move) int {
	var result = (move.MovingPiece() << 6) | move.To()
	if side {
		result |= 1 << 9
	}
	return result
}
