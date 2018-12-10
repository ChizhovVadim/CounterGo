package engine

import (
	"errors"
	"sync"

	. "github.com/ChizhovVadim/CounterGo/common"
)

var searchTimeout = errors.New("search timeout")

func (e *Engine) iterativeDeepening() {
	defer recoverFromSearchTimeout()

	var ml = e.genRootMoves()
	if len(ml) != 0 {
		e.mainLine.update(0, 0, []Move{ml[0]})
	}
	if len(ml) <= 1 {
		return
	}

	var prevScore int
	var prevBestMove Move
	for depth := 1; depth <= maxHeight; depth++ {
		e.searchRootParallel(ml, depth)
		if isDone(e.done) {
			break
		}
		if e.mainLine.score >= winIn(depth-3) ||
			e.mainLine.score <= lossIn(depth-3) {
			break
		}
		if e.timeManager.IsSoftTimeout(e.mainLine.moves[0] != prevBestMove,
			e.mainLine.score < prevScore-PawnValue/2) {
			break
		}
		prevScore = e.mainLine.score
		prevBestMove = e.mainLine.moves[0]
		e.sendProgress()
		savePV(e.transTable, &e.threads[0].stack[0].position, e.mainLine.moves)
	}
}

func savePV(transTable TransTable, p *Position, pv []Move) {
	var parent = *p
	var child Position
	for _, m := range pv {
		transTable.Update(&parent, 0, 0, 0, m)
		parent.MakeMove(m, &child)
		parent = child
	}
}

func (e *Engine) searchRootParallel(ml []Move, depth int) int {
	var mainThread = &e.threads[0]
	const height = 0
	var p = &mainThread.stack[height].position
	var alpha = -valueInfinity
	const beta = valueInfinity
	var bestMoveIndex = 0
	{
		var child = &mainThread.stack[height+1].position
		var move = ml[0]
		p.MakeMove(move, child)
		var newDepth = mainThread.newDepth(depth, height)
		var score = -mainThread.alphaBeta(-beta, -alpha, newDepth, height+1)
		alpha = score
		e.mainLine.update(depth, score,
			append([]Move{move}, mainThread.stack[height+1].pv.moves()...))
	}
	var gate = &sync.Mutex{}
	var index = 1
	parallelDo(e.Threads.Value, func(threadIndex int) {
		defer recoverFromSearchTimeout()
		var t = &e.threads[threadIndex]
		var child = &t.stack[height+1].position
		for {
			gate.Lock()
			var localAlpha = alpha
			var localIndex = index
			index++
			gate.Unlock()
			if localIndex >= len(ml) {
				return
			}
			var move = ml[localIndex]
			p.MakeMove(move, child)
			var newDepth = t.newDepth(depth, height)
			if -t.alphaBeta(-(localAlpha+1), -localAlpha, newDepth, height+1) <= localAlpha {
				continue
			}
			var score = -t.alphaBeta(-beta, -localAlpha, newDepth, height+1)
			gate.Lock()
			if score > alpha {
				alpha = score
				e.mainLine.update(depth, score,
					append([]Move{move}, t.stack[height+1].pv.moves()...))
				bestMoveIndex = localIndex
			}
			gate.Unlock()
		}
	})
	moveToBegin(ml, bestMoveIndex)
	return alpha
}

func (t *thread) alphaBeta(alpha, beta, depth, height int) int {
	var newDepth, score int
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

	var hashMove = MoveEmpty
	var pvNode = beta != alpha+1

	if ttDepth, ttScore, ttType, ttMove, ok := t.engine.transTable.Read(position); ok {
		hashMove = ttMove
		if ttDepth >= depth && !pvNode {
			ttScore = valueFromTT(ttScore, height)
			if ttScore >= beta && (ttType&boundLower) != 0 {
				return beta
			}
			if ttScore <= alpha && (ttType&boundUpper) != 0 {
				return alpha
			}
		}
	}

	var lazyEval = lazyEval{evaluator: t.evaluator, position: position}

	var child = &t.stack[height+1].position
	if depth >= 2 && !pvNode && !isCheck && position.LastMove != MoveEmpty &&
		beta < valueWin &&
		!isLateEndgame(position, position.WhiteMove) &&
		(depth-4 <= 0 || lazyEval.Value() >= beta) {
		newDepth = depth - 4
		position.MakeNullMove(child)
		if newDepth <= 0 {
			score = -t.quiescence(-beta, -(beta - 1), 1, height+1)
		} else {
			score = -t.alphaBeta(-beta, -(beta - 1), newDepth, height+1)
		}
		if score >= beta && score < valueWin {
			return beta
		}
	}

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

			newDepth = t.newDepth(depth, height)
			var reduction = 0

			if !isCaptureOrPromotion(move) && moveCount > 1 &&
				!isCheck && !child.IsCheck() &&
				ml[i].Key < sortTableKeyImportant &&
				!isPawnAdvance(move, position.WhiteMove) &&
				alpha > valueLoss {

				if depth >= 3 {
					reduction = t.engine.lateMoveReduction(depth, moveCount)
					if pvNode {
						reduction--
					}
					reduction = Max(0, Min(depth-2, reduction))
				} else {
					if position.LastMove != MoveEmpty &&
						moveCount >= 9+3*depth {
						continue
					}
					if position.LastMove != MoveEmpty &&
						lazyEval.Value()+PawnValue*depth <= alpha {
						continue
					}
				}
			}

			if !isCaptureOrPromotion(move) {
				quietsSearched = append(quietsSearched, move)
			}

			if reduction > 0 ||
				(beta != alpha+1 && moveCount > 1 && newDepth > 0) {
				score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth-reduction, height+1)
				if score <= alpha {
					continue
				}
			}

			score = -t.alphaBeta(-beta, -alpha, newDepth, height+1)

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

func (t *thread) quiescence(alpha, beta, depth, height int) int {
	t.stack[height].pv.clear()
	t.incNodes()
	if height >= maxHeight {
		return valueDraw
	}
	var position = &t.stack[height].position
	var isCheck = position.IsCheck()
	var eval = 0
	if !isCheck {
		eval = t.evaluator.Evaluate(position)
		if eval > alpha {
			alpha = eval
		}
		if eval >= beta {
			return alpha
		}
	}
	var ml = t.stack[height].moveList[:]
	if position.IsCheck() {
		ml = position.GenerateMoves(ml)
	} else {
		ml = position.GenerateCaptures(ml, depth > 0)
	}
	t.sortTable.NoteQS(position, ml)
	sortMoves(ml)
	var moveCount = 0
	var child = &t.stack[height+1].position
	for i := range ml {
		var move = ml[i].Move
		var danger = isDangerCapture(position, move)
		if !isCheck && !danger && !seeGEZero(position, move) {
			continue
		}
		if position.MakeMove(move, child) {
			moveCount++
			if !isCheck && !danger && !child.IsCheck() &&
				eval+moveValue(move)+2*PawnValue <= alpha {
				continue
			}
			var score = -t.quiescence(-beta, -alpha, depth-1, height+1)
			if score > alpha {
				alpha = score
				if score >= beta {
					break
				}
				t.stack[height].pv.assign(move, &t.stack[height+1].pv)
			}
		}
	}
	if isCheck && moveCount == 0 {
		return lossIn(height)
	}
	return alpha
}

func (t *thread) incNodes() {
	t.nodes++
	if (t.nodes&255) == 0 && isDone(t.engine.done) {
		panic(searchTimeout)
	}
}

func isDone(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

func (t *thread) isDraw(height int) bool {
	var p = &t.stack[height].position

	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	if p.Rule50 > 100 {
		return true
	}

	for i := height - 1; i >= 0; i-- {
		var temp = &t.stack[i].position
		if temp.Key == p.Key {
			return true
		}
		if temp.Rule50 == 0 || temp.LastMove == MoveEmpty {
			return false
		}
	}

	if t.engine.historyKeys[p.Key] >= 2 {
		return true
	}

	return false
}

func (t *thread) newDepth(depth, height int) int {
	var p = &t.stack[height].position
	var child = &t.stack[height+1].position
	var prevMove = p.LastMove
	var move = child.LastMove
	var givesCheck = child.IsCheck()

	if prevMove != MoveEmpty &&
		prevMove.To() == move.To() &&
		move.CapturedPiece() > Pawn &&
		prevMove.CapturedPiece() > Pawn &&
		seeGEZero(p, move) {
		return depth
	}

	if givesCheck && (depth <= 1 || seeGEZero(p, move)) {
		return depth
	}

	if isPawnPush7th(move, p.WhiteMove) && seeGEZero(p, move) {
		return depth
	}

	return depth - 1
}

func recoverFromSearchTimeout() {
	var r = recover()
	if r != nil && r != searchTimeout {
		panic(r)
	}
}

func moveToBegin(ml []Move, index int) {
	if index == 0 {
		return
	}
	var item = ml[index]
	for i := index; i > 0; i-- {
		ml[i] = ml[i-1]
	}
	ml[0] = item
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

func (e *Engine) genRootMoves() []Move {
	var t = e.threads[0]
	const height = 0
	var p = &t.stack[height].position
	_, _, _, transMove, _ := e.transTable.Read(p)
	var ml = p.GenerateMoves(t.stack[height].moveList[:])
	t.sortTable.Note(p, ml, transMove, height)
	sortMoves(ml)
	var result []Move
	var child = &t.stack[height+1].position
	for i := range ml {
		var move = ml[i].Move
		if p.MakeMove(move, child) {
			result = append(result, move)
		}
	}
	return result
}

type lazyEval struct {
	evaluator Evaluator
	position  *Position
	hasValue  bool
	value     int
}

func (le *lazyEval) Value() int {
	if !le.hasValue {
		le.value = le.evaluator.Evaluate(le.position)
		le.hasValue = true
	}
	return le.value
}
