package engine

import "sync/atomic"

type historyTable []int32

func NewHistoryTable() historyTable {
	return make([]int32, 1<<10)
}

func (ht historyTable) Clear() {
	for i := 0; i < len(ht); i++ {
		ht[i] = 0
	}
}

func (ht historyTable) Update(ctx *searchContext, bestMove Move, depth int) {
	if ctx.Killer1 != bestMove {
		ctx.Killer2 = ctx.Killer1
		ctx.Killer1 = bestMove
	}
	var side = ctx.Position.WhiteMove
	atomic.AddInt32(&ht[pieceSquareIndex(side, bestMove)], int32(len(ctx.QuietsSearched)*depth))
	for _, move := range ctx.QuietsSearched {
		atomic.AddInt32(&ht[pieceSquareIndex(side, move)], int32(-depth))
	}
}

func (ht historyTable) Score(side bool, move Move) int {
	return int(ht[pieceSquareIndex(side, move)])
}

func pieceSquareIndex(side bool, move Move) int {
	var result = (move.MovingPiece() << 6) | move.To()
	if side {
		result |= 1 << 9
	}
	return result
}
