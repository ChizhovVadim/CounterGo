package engine

import "sync/atomic"

type historyEntry struct {
	success, try int32
}

type HistoryTable []historyEntry

func NewHistoryTable() HistoryTable {
	return make([]historyEntry, 7*2*64)
}

func (ht HistoryTable) historyEntry(side bool, move Move) *historyEntry {
	var index = MakePiece(move.MovingPiece(), side)*64 + move.To()
	return &ht[index]
}

func (ht HistoryTable) Clear() {
	for i := 0; i < len(ht); i++ {
		ht[i] = historyEntry{0, 0}
	}
}

func (ht HistoryTable) Update(ss *SearchStack, bestMove Move, depth int) {
	ss.KillerMove = bestMove
	var side = ss.Position.WhiteMove
	atomic.AddInt32(&ht.historyEntry(side, bestMove).success, int32(depth))
	for _, move := range ss.QuietsSearched {
		atomic.AddInt32(&ht.historyEntry(side, move).try, int32(depth))
	}
}

func (ht HistoryTable) Score(side bool, move Move) int {
	var entry = ht.historyEntry(side, move)
	if entry.try == 0 {
		return 0
	}
	return int(100 * entry.success / entry.try)
}
