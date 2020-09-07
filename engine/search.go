package engine

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const pawnValue = 100

var errSearchTimeout = errors.New("search timeout")

/*func savePV(transTable TransTable, p *Position, pv []Move) {
	var parent = *p
	var child Position
	for _, m := range pv {
		transTable.Update(&parent, 0, 0, 0, m)
		parent.MakeMove(m, &child)
		parent = child
	}
}*/

func lazySmp(ctx context.Context, e *Engine) {
	var ml = e.genRootMoves()
	if len(ml) != 0 {
		e.mainLine = mainLine{
			depth: 0,
			score: 0,
			moves: []Move{ml[0]},
		}
		e.depth = 0
	}
	if len(ml) <= 1 {
		return
	}

	e.done = ctx.Done()

	if e.Threads == 1 {
		iterativeDeepening(&e.threads[0], ml, 1, 1)
	} else {

		var wg = &sync.WaitGroup{}

		for i := 0; i < e.Threads; i++ {
			var ml = cloneMoves(ml)
			wg.Add(1)
			go func(i int) {
				var t = &e.threads[i]
				iterativeDeepening(t, ml, 1+i%2, 2)
				wg.Done()
			}(i)
		}

		wg.Wait()
	}
}

func iterativeDeepening(t *thread, ml []Move, startDepth, incDepth int) { //TODO, aspirationMargin
	for depth := startDepth; depth <= maxHeight; depth += incDepth {
		t.depth = int32(depth)
		if isDone(t.engine.done) {
			break
		}

		var globalLine mainLine
		t.engine.mu.Lock()
		globalLine = t.engine.mainLine
		t.engine.mu.Unlock()

		if depth <= globalLine.depth {
			continue
		}
		if index := findMoveIndex(ml, globalLine.moves[0]); index >= 0 {
			moveToBegin(ml, index)
		}

		var score, iterationComplete = aspirationWindow(t, ml, depth, globalLine.score)
		if iterationComplete {
			t.engine.mu.Lock()
			if depth > t.engine.mainLine.depth {
				atomic.StoreInt32(&t.engine.depth, int32(depth))
				t.engine.mainLine = mainLine{
					depth: depth,
					score: score,
					moves: t.stack[0].pv.toSlice(),
				}
				t.engine.timeManager.OnIterationComplete(t.engine.mainLine)
				t.engine.sendProgress()
			}
			t.engine.mu.Unlock()
		}
	}
}

func aspirationWindow(t *thread, ml []Move, depth, prevScore int) (int, bool) {
	defer recoverFromSearchTimeout()
	if depth >= 5 && !(prevScore <= valueLoss || prevScore >= valueWin) {
		var alphaMargin = 25
		var betaMargin = 25
		for i := 0; i < 2; i++ {
			var alpha = Max(-valueInfinity, prevScore-alphaMargin)
			var beta = Min(valueInfinity, prevScore+betaMargin)
			var score = searchRoot(t, ml, alpha, beta, depth)
			if score >= valueWin || score <= valueLoss {
				break
			} else if score >= beta {
				betaMargin *= 2
			} else if score <= alpha {
				alphaMargin *= 2
			} else {
				return score, true
			}
		}
	}
	return searchRoot(t, ml, -valueInfinity, valueInfinity, depth), true
}

func searchRoot(t *thread, ml []Move, alpha, beta, depth int) int {
	const height = 0
	t.stack[height].pv.clear()
	var p = &t.stack[height].position
	t.stack[height].staticEval = t.evaluator.Evaluate(p)
	var child = &t.stack[height+1].position
	var bestMoveIndex = 0
	for i, move := range ml {
		p.MakeMove(move, child)
		var extension, reduction int
		extension = t.extend(depth, height)
		if depth >= 3 && i > 0 &&
			!(isCaptureOrPromotion(move)) {
			reduction = t.engine.lateMoveReduction(depth, i+1)
			reduction = Max(0, Min(depth-2, reduction))
		}
		var newDepth = depth - 1 + extension
		var nextFirstline = i == 0
		if (reduction > 0) &&
			-t.alphaBeta(-(alpha+1), -alpha, newDepth-reduction, height+1, nextFirstline) <= alpha {
			continue
		}
		var score = -t.alphaBeta(-beta, -alpha, newDepth, height+1, nextFirstline)
		if score > alpha {
			alpha = score
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
			bestMoveIndex = i
			if alpha >= beta {
				break
			}
		}
	}
	moveToBegin(ml, bestMoveIndex)
	return alpha
}

func (t *thread) alphaBeta(alpha, beta, depth, height int, firstline bool) int {
	var newDepth, score int
	t.stack[height].pv.clear()

	var position = &t.stack[height].position

	if height >= maxHeight {
		return t.evaluator.Evaluate(position)
	}

	if t.isDraw(height) {
		return valueDraw
	}

	if depth <= 0 {
		return t.quiescence(alpha, beta, 1, height)
	}

	t.incNodes()

	var isCheck = position.IsCheck()

	// mate distance pruning
	if winIn(height+1) <= alpha {
		return alpha
	}
	if lossIn(height+2) >= beta && !isCheck {
		return beta
	}

	// transposition table
	var ttDepth, ttValue, ttBound, ttMove, ttHit = t.engine.transTable.Read(position)
	if ttHit {
		ttValue = valueFromTT(ttValue, height)
		if ttDepth >= depth {
			if ttValue >= beta && (ttBound&boundLower) != 0 {
				if ttMove != MoveEmpty && !isCaptureOrPromotion(ttMove) {
					t.sortTable.Update(position, ttMove, nil, depth, height)
				}
				return beta
			}
			if ttValue <= alpha && (ttBound&boundUpper) != 0 {
				return alpha
			}
		}
	}

	var staticEval = t.evaluator.Evaluate(position)
	t.stack[height].staticEval = staticEval
	var improving = height >= 2 && staticEval > t.stack[height-2].staticEval

	// reverse futility pruning
	if !firstline && depth <= 5 && !isCheck &&
		beta < valueWin && beta > valueLoss &&
		staticEval-pawnValue*depth >= beta {
		return beta
	}

	// null-move pruning
	var child = &t.stack[height+1].position
	if !firstline && depth >= 3 && !isCheck && position.LastMove != MoveEmpty &&
		beta < valueWin && beta > valueLoss &&
		!(ttHit && ttValue < beta && (ttBound&boundUpper) != 0) &&
		!isLateEndgame(position, position.WhiteMove) &&
		staticEval >= beta {
		var reduction = 4 + (depth-2)/6
		if staticEval >= beta+2*pawnValue {
			reduction++
		}
		reduction = Min(reduction, depth-1)
		if reduction >= 2 {
			position.MakeNullMove(child)
			score = -t.alphaBeta(-beta, -(beta - 1), depth-reduction, height+1, false)
			if score >= beta {
				return beta
			}
		}
	}

	var ml = position.GenerateMoves(t.stack[height].moveList[:])
	t.sortTable.Note(position, ml, ttMove, height)

	// singular extension
	var ttMoveIsSingular = false
	if depth >= 8 &&
		ttHit && ttMove != MoveEmpty &&
		(ttBound&boundLower) != 0 && ttDepth >= depth-3 &&
		ttValue > valueLoss && ttValue < valueWin {

		ttMoveIsSingular = true
		sortMoves(ml)
		var singularBeta = Max(-valueInfinity, ttValue-pawnValue/2)
		newDepth = depth/2 - 1
		for i := range ml {
			var move = ml[i].Move
			if !position.MakeMove(move, child) {
				continue
			}
			if move == ttMove {
				if t.extend(depth, height) == 1 {
					ttMoveIsSingular = false
					break
				}
				continue
			}
			score = -t.alphaBeta(-singularBeta, -singularBeta+1, newDepth, height+1, false)
			if score >= singularBeta {
				ttMoveIsSingular = false
				break
			}
		}

	}

	var moveCount = 0
	var quietsSearched = t.stack[height].quietsSearched[:0]
	var bestMove Move
	const SortMovesIndex = 4

	var lmp = 5 + depth*depth
	if !improving {
		lmp /= 2
	}

	for i := range ml {
		if i < SortMovesIndex {
			moveToTop(ml[i:])
		} else if i == SortMovesIndex {
			sortMoves(ml[i:])
		}
		var move = ml[i].Move

		if !position.MakeMove(move, child) {
			continue
		}
		moveCount++

		if !(alpha <= valueLoss ||
			ml[i].Key >= sortTableKeyImportant ||
			isCheck ||
			child.IsCheck() ||
			isCaptureOrPromotion(move) ||
			isPawnAdvance(move, position.WhiteMove)) {
			// late-move pruning
			if lmp != -1 && moveCount > lmp {
				continue
			}
		}

		// SEE pruning
		if depth <= 4 &&
			!(alpha <= valueLoss ||
				isCheck ||
				isCaptureOrPromotion(move) ||
				move == ttMove ||
				move.MovingPiece() == King) &&
			!seeGEZero(position, move) {
			continue
		}

		var extension, reduction int

		extension = t.extend(depth, height)
		if move == ttMove && ttMoveIsSingular {
			extension = 1
		}

		if depth >= 3 && moveCount > 1 &&
			!(ml[i].Key >= sortTableKeyImportant ||
				isCaptureOrPromotion(move)) {
			reduction = t.engine.lateMoveReduction(depth, moveCount)
			reduction = Max(0, Min(depth-2, reduction))
		}

		if !isCaptureOrPromotion(move) {
			quietsSearched = append(quietsSearched, move)
		}

		newDepth = depth - 1 + extension
		var nextFirstline = firstline && moveCount == 1

		// LMR
		if reduction > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth-reduction, height+1, nextFirstline)
			if score <= alpha {
				continue
			}
		}

		score = -t.alphaBeta(-beta, -alpha, newDepth, height+1, nextFirstline)

		if score > alpha {
			alpha = score
			bestMove = move
			if alpha >= beta {
				break
			}
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
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

	ttBound = 0
	if bestMove != MoveEmpty {
		ttBound |= boundLower
	}
	if alpha < beta {
		ttBound |= boundUpper
	}
	t.engine.transTable.Update(position, depth, valueToTT(alpha, height), ttBound, bestMove)

	return alpha
}

func (t *thread) quiescence(alpha, beta, depth, height int) int {
	t.stack[height].pv.clear()
	t.incNodes()
	var position = &t.stack[height].position
	if height >= maxHeight {
		return t.evaluator.Evaluate(position)
	}
	var isCheck = position.IsCheck()
	if !isCheck {
		var eval = t.evaluator.Evaluate(position)
		if eval > alpha {
			alpha = eval
			if alpha >= beta {
				return alpha
			}
		}
	}
	var ml = t.stack[height].moveList[:]
	if isCheck {
		ml = position.GenerateMoves(ml)
	} else {
		ml = position.GenerateCaptures(ml)
	}
	t.sortTable.NoteQS(position, ml)
	sortMoves(ml)
	var moveCount = 0
	var child = &t.stack[height+1].position
	for i := range ml {
		var move = ml[i].Move
		if !isCheck && !seeGEZero(position, move) {
			continue
		}
		if !position.MakeMove(move, child) {
			continue
		}
		moveCount++
		var score = -t.quiescence(-beta, -alpha, depth-1, height+1)
		if score > alpha {
			alpha = score
			if alpha >= beta {
				break
			}
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
		}
	}
	if isCheck && moveCount == 0 {
		return lossIn(height)
	}
	return alpha
}

func (t *thread) incNodes() {
	t.nodes++
	if t.nodes > 255 {
		var globalNodes = atomic.AddInt64(&t.engine.nodes, t.nodes)
		var globalDepth = atomic.LoadInt32(&t.engine.depth)
		t.nodes = 0
		t.engine.timeManager.OnNodesChanged(int(globalNodes))
		if t.depth <= globalDepth ||
			isDone(t.engine.done) {
			panic(errSearchTimeout)
		}
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

	if p.Rule50 == 0 || p.LastMove == MoveEmpty {
		return false
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

func (t *thread) extend(depth, height int) int {
	var child = &t.stack[height+1].position
	var givesCheck = child.IsCheck()

	if givesCheck {
		return 1
	}

	return 0
}

func recoverFromSearchTimeout() {
	var r = recover()
	if r != nil && r != errSearchTimeout {
		panic(r)
	}
}

func findMoveIndex(ml []Move, move Move) int {
	for i := range ml {
		if ml[i] == move {
			return i
		}
	}
	return -1
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

func cloneMoves(ml []Move) []Move {
	var result = make([]Move, len(ml))
	copy(result, ml)
	return result
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
