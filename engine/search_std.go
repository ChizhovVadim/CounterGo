package engine

import (
	. "github.com/ChizhovVadim/CounterGo/common"
)

func (e *Engine) searchRootStd(ml []Move, depth int) int {
	var t = &e.threads[0]
	const height = 0
	var p = &t.stack[height].position
	var child = &t.stack[height+1].position
	var alpha = -valueInfinity
	const beta = valueInfinity
	var bestMoveIndex = 0
	for i, move := range ml {
		p.MakeMove(move, child)
		var newDepth = t.newDepth(depth, height)
		var score = -t.alphaBetaStd(-beta, -alpha, newDepth, height+1)
		if score > alpha {
			alpha = score
			bestMoveIndex = i
			e.mainLine.update(depth, score,
				append([]Move{move}, t.stack[height+1].pv.moves()...))
		}
	}
	moveToBegin(ml, bestMoveIndex)
	return alpha
}

func (t *thread) alphaBetaStd(alpha, beta, depth, height int) int {
	t.stack[height].pv.clear()

	if height >= maxHeight || t.isDraw(height) {
		return valueDraw
	}

	if depth <= 0 {
		return t.quiescence(alpha, beta, 1, height)
	}

	t.incNodes()

	var position = &t.stack[height].position
	var isCheck = position.IsCheck()

	if winIn(height+1) <= alpha {
		return alpha
	}
	if lossIn(height+2) >= beta && !isCheck {
		return beta
	}

	_, _, _, hashMove, _ := t.engine.transTable.Read(position)

	var child = &t.stack[height+1].position

	var ml = position.GenerateMoves(t.stack[height].moveList[:])
	t.sortTable.Note(position, ml, hashMove, height)

	var moveCount = 0
	var quietsSearched = t.stack[height].quietsSearched[:0]
	var bestMove Move
	const SortMovesIndex = 4

	for i := range ml {
		if i < SortMovesIndex {
			moveToTop(ml[i:])
		} else if i == SortMovesIndex {
			sortMoves(ml[i:])
		}
		var move = ml[i].Move

		if position.MakeMove(move, child) {
			moveCount++

			var newDepth = t.newDepth(depth, height)

			if !isCaptureOrPromotion(move) {
				quietsSearched = append(quietsSearched, move)
			}

			var score = -t.alphaBetaStd(-beta, -alpha, newDepth, height+1)

			if score > alpha {
				alpha = score
				bestMove = move
				if alpha >= beta {
					break
				}
				t.stack[height].pv.assign(move, &t.stack[height+1].pv)
			}
		}
	}

	if moveCount == 0 {
		if isCheck {
			return lossIn(height)
		}
		return valueDraw
	}

	if bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		t.sortTable.Update(position, bestMove, quietsSearched, depth, height)
	}

	var ttType = 0
	if bestMove != MoveEmpty {
		ttType |= boundLower
	}
	if alpha < beta {
		ttType |= boundUpper
	}
	t.engine.transTable.Update(position, depth, valueToTT(alpha, height), ttType, bestMove)

	return alpha
}
