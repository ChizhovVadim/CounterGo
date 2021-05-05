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
var errSearchLowDepth = errors.New("search low depth")

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
	defer func() {
		if r := recover(); r != nil {
			if r == errSearchTimeout {
				return
			}
			panic(r)
		}
	}()

	const height = 0
	t.sortTable.ResetKillers(height)
	t.sortTable.ResetKillers(height + 1)
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
			t.engine.updateMainLine(mainLine{
				depth: depth,
				score: score,
				moves: t.stack[height].pv.toSlice(),
			})
		}
	}
}

func aspirationWindow(t *thread, ml []Move, depth, prevScore int) (int, bool) {
	defer func() {
		if r := recover(); r != nil {
			if r == errSearchLowDepth {
				return
			}
			panic(r)
		}
	}()

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

// main search method
func (t *thread) alphaBeta(alpha, beta, depth, height int, firstline bool) int {
	var oldAlpha = alpha
	var newDepth, score int
	t.stack[height].pv.clear()

	var position = &t.stack[height].position

	if height >= maxHeight {
		return t.evaluator.Evaluate(position)
	}

	if repetition, inCurrentSearch := t.isRepeat(height); repetition {
		alpha = Max(alpha, valueDraw)
		if alpha >= beta || inCurrentSearch {
			return alpha
		}
	}

	if depth <= 0 {
		return t.quiescence(alpha, beta, height)
	}

	t.incNodes()

	if isDraw(position) {
		return valueDraw
	}

	var isCheck = position.IsCheck()

	// mate distance pruning
	if winIn(height+1) <= alpha {
		return alpha
	}
	if lossIn(height+2) >= beta && !isCheck {
		return beta
	}

	// transposition table
	var ttDepth, ttValue, ttBound, ttMove, ttHit = t.engine.transTable.Read(position.Key)
	if ttHit {
		ttValue = valueFromTT(ttValue, height)
		if ttDepth >= depth {
			if ttValue >= beta && (ttBound&boundLower) != 0 {
				if ttMove != MoveEmpty && !isCaptureOrPromotion(ttMove) {
					t.sortTable.Update(position, ttMove, nil, depth, height)
				}
				return ttValue
			}
			if ttValue <= alpha && (ttBound&boundUpper) != 0 {
				return ttValue
			}
		}
	}

	var staticEval = t.evaluator.Evaluate(position)
	t.stack[height].staticEval = staticEval
	var improving = position.LastMove == MoveEmpty ||
		height >= 2 && staticEval > t.stack[height-2].staticEval

	// reverse futility pruning
	if !firstline && depth <= 8 && !isCheck {
		var score = staticEval - pawnValue*depth
		if score >= beta {
			return beta
		}
	}

	t.sortTable.ResetKillers(height + 1)

	// null-move pruning
	var child = &t.stack[height+1].position
	if !firstline && depth >= 2 && !isCheck &&
		position.LastMove != MoveEmpty &&
		(height <= 1 || t.stack[height-1].position.LastMove != MoveEmpty) &&
		beta < valueWin &&
		!(ttHit && ttValue < beta && (ttBound&boundUpper) != 0) &&
		!isLateEndgame(position, position.WhiteMove) &&
		staticEval >= beta {
		var skepticStaticEval = Min(staticEval, evalMaterial(position))
		var reduction = 4 + depth/6 + Max(0, Min(3, (skepticStaticEval-beta)/200))
		position.MakeNullMove(child)
		score = -t.alphaBeta(-beta, -(beta - 1), depth-reduction, height+1, false)
		if score >= beta {
			return beta
		}
	}

	// Internal iterative deepening
	if depth >= 8 && ttMove == MoveEmpty && !isCheck {
		t.alphaBeta(alpha, beta, depth-7, height, firstline)
		if t.stack[height].pv.size != 0 {
			ttMove = t.stack[height].pv.items[0]
			t.stack[height].pv.clear()
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
		var singularBeta = Max(-valueInfinity, ttValue-2*depth)
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

	var movesSearched = 0
	var hasLegalMove = false
	var movesSeen = 0

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
		movesSeen++

		if depth <= 8 && alpha > valueLoss && hasLegalMove {
			// late-move pruning
			if !isCaptureOrPromotion(move) && movesSeen > lmp {
				continue
			}

			// SEE pruning
			if !isCheck && staticEval-pawnValue*depth <= alpha &&
				!SeeGE(position, move, -depth) {
				continue
			}
		}

		if !position.MakeMove(move, child) {
			movesSeen--
			continue
		}
		hasLegalMove = true

		movesSearched++

		var extension, reduction int

		extension = t.extend(depth, height)
		if move == ttMove && ttMoveIsSingular {
			extension = 1
		}

		if depth >= 3 && movesSearched > 1 &&
			!(isCaptureOrPromotion(move)) {
			reduction = t.engine.lateMoveReduction(depth, movesSearched)
			if ml[i].Key >= sortTableKeyImportant {
				reduction--
			}
			reduction = Max(0, Min(depth-2, reduction))
		}

		if !isCaptureOrPromotion(move) {
			quietsSearched = append(quietsSearched, move)
		}

		newDepth = depth - 1 + extension
		var nextFirstline = firstline && movesSearched == 1

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
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
			if alpha >= beta {
				break
			}
		}
	}

	if !hasLegalMove {
		if isCheck {
			return lossIn(height)
		}
		return valueDraw
	}

	if bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		t.sortTable.Update(position, bestMove, quietsSearched, depth, height)
	}

	ttBound = 0
	if alpha > oldAlpha {
		ttBound |= boundLower
	}
	if alpha < beta {
		ttBound |= boundUpper
	}
	t.engine.transTable.Update(position.Key, depth, valueToTT(alpha, height), ttBound, bestMove)

	return alpha
}

func (t *thread) quiescence(alpha, beta, height int) int {
	t.stack[height].pv.clear()
	t.incNodes()
	var position = &t.stack[height].position
	if isDraw(position) {
		return valueDraw
	}
	if height >= maxHeight {
		return t.evaluator.Evaluate(position)
	}

	var _, ttValue, ttBound, _, ttHit = t.engine.transTable.Read(position.Key)
	if ttHit {
		ttValue = valueFromTT(ttValue, height)
		if ttBound == boundExact ||
			ttBound == boundLower && ttValue >= beta ||
			ttBound == boundUpper && ttValue <= alpha {
			return ttValue
		}
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
	var hasLegalMove = false
	var child = &t.stack[height+1].position
	for i := range ml {
		var move = ml[i].Move
		if !isCheck && !seeGEZero(position, move) {
			continue
		}
		if !position.MakeMove(move, child) {
			continue
		}
		hasLegalMove = true
		var score = -t.quiescence(-beta, -alpha, height+1)
		if score > alpha {
			alpha = score
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
			if alpha >= beta {
				break
			}
		}
	}
	if isCheck && !hasLegalMove {
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
		if t.depth <= globalDepth {
			panic(errSearchLowDepth)
		}
		if isDone(t.engine.done) {
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

func isDraw(p *Position) bool {
	if p.Rule50 > 100 {
		return true
	}

	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	return false
}

func (t *thread) isRepeat(height int) (repetition, inCurrentSearch bool) {
	var p = &t.stack[height].position

	if p.Rule50 == 0 {
		return false, false
	}
	for i := height - 1; i >= 0; i-- {
		var temp = &t.stack[i].position
		if temp.Key == p.Key {
			return true, true
		}
		if temp.Rule50 == 0 || temp.LastMove == MoveEmpty {
			return false, false
		}
	}

	if t.engine.historyKeys[p.Key] >= 2 {
		return true, false
	}

	return false, false
}

func (t *thread) extend(depth, height int) int {
	var p = &t.stack[height].position
	var child = &t.stack[height+1].position
	var givesCheck = child.IsCheck()

	if givesCheck && (depth <= 2 || seeGEZero(p, child.LastMove)) {
		return 1
	}

	return 0
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
	var t = &e.threads[0]
	const height = 0
	var p = &t.stack[height].position
	_, _, _, transMove, _ := e.transTable.Read(p.Key)
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
