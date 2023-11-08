package engine

import . "github.com/ChizhovVadim/CounterGo/pkg/common"

const historyMax = 1 << 14

type historyContext struct {
	thread     *thread
	sideToMove bool
	cont1      int
	cont2      int
}

func (h *historyContext) ReadTotal(m Move) int {
	var sideToMove = h.sideToMove
	var score int
	score += int(h.thread.mainHistory[sideFromToIndex(sideToMove, m)])
	var pieceToIndex = pieceSquareIndex(sideToMove, m)
	if h.cont1 != -1 {
		score += int(h.thread.continuationHistory[h.cont1][pieceToIndex])
	}
	if h.cont2 != -1 {
		score += int(h.thread.continuationHistory[h.cont2][pieceToIndex])
	}
	return score
}

func (h *historyContext) Update(quietsSearched []Move, bestMove Move, depth int) {
	var bonus = Min(depth*depth, 400)
	var t = h.thread
	var sideToMove = h.sideToMove
	var cont1 = h.cont1
	var cont2 = h.cont2

	for _, m := range quietsSearched {
		var good = m == bestMove

		var fromToIndex = sideFromToIndex(sideToMove, m)
		updateHistory(&t.mainHistory[fromToIndex], bonus, good)
		var pieceToIndex = pieceSquareIndex(sideToMove, m)
		if cont1 != -1 {
			updateHistory(&t.continuationHistory[cont1][pieceToIndex], bonus, good)
		}
		if cont2 != -1 {
			updateHistory(&t.continuationHistory[cont2][pieceToIndex], bonus, good)
		}

		if good {
			break
		}
	}
}

// Exponential moving average
func updateHistory(v *int16, bonus int, good bool) {
	var newVal int
	if good {
		newVal = historyMax
	} else {
		newVal = -historyMax
	}
	*v += int16((newVal - int(*v)) * bonus / 512)
}

func (t *thread) clearHistory() {
	for i := range t.mainHistory {
		t.mainHistory[i] = 0
	}
	for i := range t.continuationHistory {
		for j := range t.continuationHistory[i] {
			t.continuationHistory[i][j] = 0
		}
	}
}

func (t *thread) getHistoryContext(height int) historyContext {
	var sideToMove = t.stack[height].position.WhiteMove
	var cont1 = -1
	{
		var prev1 = t.stack[height].position.LastMove
		if prev1 != MoveEmpty {
			cont1 = pieceSquareIndex(!sideToMove, prev1)
		}
	}
	var cont2 = -1
	if height > 0 {
		var prev2 = t.stack[height-1].position.LastMove
		if prev2 != MoveEmpty {
			cont2 = pieceSquareIndex(sideToMove, prev2)
		}
	}
	return historyContext{
		thread:     t,
		sideToMove: sideToMove,
		cont1:      cont1,
		cont2:      cont2,
	}
}

func pieceSquareIndex(side bool, move Move) int {
	var result = (move.MovingPiece() << 6) | move.To()
	if side {
		result |= 1 << 9
	}
	return result
}

func sideFromToIndex(side bool, move Move) int {
	var result = (move.From() << 6) | move.To()
	if side {
		result |= 1 << 12
	}
	return result
}
