package engine

import "sync/atomic"

type historyEntry struct {
	success, try int32
}

type historyTable []historyEntry

func NewHistoryTable() historyTable {
	return make([]historyEntry, 7*2*64)
}

func (ht historyTable) historyEntry(side bool, move Move) *historyEntry {
	var index = MakePiece(move.MovingPiece(), side)*64 + move.To()
	return &ht[index]
}

func (ht historyTable) Clear() {
	for i := 0; i < len(ht); i++ {
		ht[i] = historyEntry{0, 0}
	}
}

func (ht historyTable) Update(ctx *searchContext, bestMove Move, depth int) {
	if ctx.Killer1 != bestMove {
		ctx.Killer2 = ctx.Killer1
		ctx.Killer1 = bestMove
	}
	var side = ctx.Position.WhiteMove
	atomic.AddInt32(&ht.historyEntry(side, bestMove).success, int32(depth))
	for _, move := range ctx.QuietsSearched {
		atomic.AddInt32(&ht.historyEntry(side, move).try, int32(depth))
	}
}

func (ht historyTable) Score(side bool, move Move) int {
	var entry = ht.historyEntry(side, move)
	if entry.try == 0 {
		return 0
	}
	return int(100 * entry.success / entry.try)
}
