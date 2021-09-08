package engine

import . "github.com/ChizhovVadim/CounterGo/common"

const historyMax = 1 << 14

type historyService struct {
	ButterflyHistory [8192]int16
	FollowUpHistory  [1024][1024]int16
	CounterHistory   [1024][1024]int16
}

type historyContext struct {
	Butterfly *[8192]int16
	FollowUp  *[1024]int16
	Counter   *[1024]int16
}

func (h *historyContext) ReadTotal(sideToMove bool, m Move) int {
	var score int
	if h.Butterfly != nil {
		score += int(h.Butterfly[sideFromToIndex(sideToMove, m)])
	}
	var pieceToIndex = pieceSquareIndex(sideToMove, m)
	if h.Counter != nil {
		score += int(h.Counter[pieceToIndex])
	}
	if h.FollowUp != nil {
		score += int(h.FollowUp[pieceToIndex])
	}
	return score
}

func (h *historyContext) Update(sideToMove bool, quietsSearched []Move, bestMove Move, depth int) {
	var bonus = Min(depth*depth, 400)

	for _, m := range quietsSearched {
		var newVal int
		if m == bestMove {
			newVal = historyMax
		} else {
			newVal = -historyMax
		}

		// Exponential moving average
		if h.Butterfly != nil {
			var fromToIndex = sideFromToIndex(sideToMove, m)
			h.Butterfly[fromToIndex] += int16((newVal - int(h.Butterfly[fromToIndex])) * bonus / 512)
		}
		var pieceToIndex = pieceSquareIndex(sideToMove, m)
		if h.Counter != nil {
			h.Counter[pieceToIndex] += int16((newVal - int(h.Counter[pieceToIndex])) * bonus / 512)
		}
		if h.FollowUp != nil {
			h.FollowUp[pieceToIndex] += int16((newVal - int(h.FollowUp[pieceToIndex])) * bonus / 512)
		}

		if m == bestMove {
			break
		}
	}
}

func (h *historyService) Clear() {
	for i := range h.ButterflyHistory {
		h.ButterflyHistory[i] = 0
	}
	for i := 0; i < 1024; i++ {
		for j := 0; j < 1024; j++ {
			h.FollowUpHistory[i][j] = 0
			h.CounterHistory[i][j] = 0
		}
	}
}

func (h *historyService) getContext(sideToMove bool, counter, followUp Move) historyContext {
	var result historyContext
	result.Butterfly = &h.ButterflyHistory
	if counter != MoveEmpty {
		result.Counter = &h.CounterHistory[pieceSquareIndex(sideToMove, counter)]
	}
	if followUp != MoveEmpty {
		result.FollowUp = &h.FollowUpHistory[pieceSquareIndex(sideToMove, followUp)]
	}
	return result
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
