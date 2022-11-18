package engine

import (
	"context"
	"errors"
	"sync"

	. "github.com/ChizhovVadim/CounterGo/pkg/common"
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
	}
	if len(ml) <= 1 {
		return
	}

	e.done = ctx.Done()

	if e.Threads == 1 {
		iterativeDeepening(&e.threads[0], ml, 1)
	} else {

		var wg = &sync.WaitGroup{}

		for i := 0; i < e.Threads; i++ {
			var ml = cloneMoves(ml)
			wg.Add(1)
			go func(i int) {
				var t = &e.threads[i]
				iterativeDeepening(t, ml, 1+i%2)
				wg.Done()
			}(i)
		}

		wg.Wait()
	}
}

func iterativeDeepening(t *thread, ml []Move, incDepth int) {
	defer func() {
		if r := recover(); r != nil {
			if r == errSearchTimeout {
				return
			}
			panic(r)
		}
	}()

	const height = 0
	for h := 0; h <= 2; h++ {
		t.stack[h].killer1 = MoveEmpty
		t.stack[h].killer2 = MoveEmpty
	}
	var lastDepth int
	var lastScore int
	var lastMove Move
	for {
		if isDone(t.engine.done) {
			break
		}
		var depth = lastDepth + incDepth
		if depth > maxHeight {
			break
		}
		if lastMove != MoveEmpty {
			var index = findMoveIndex(ml, lastMove)
			if index >= 0 {
				moveToBegin(ml, index)
			}
		}
		var score = aspirationWindow(t, ml, depth, lastScore)
		lastDepth, lastScore, lastMove = t.engine.onIterationComplete(t, depth, score)
	}
}

func aspirationWindow(t *thread, ml []Move, depth, prevScore int) int {
	if depth >= 5 && !(prevScore <= valueLoss || prevScore >= valueWin) {
		const Window = 25
		var alpha = Max(-valueInfinity, prevScore-Window)
		var beta = Min(valueInfinity, prevScore+Window)
		var score = searchRoot(t, ml, alpha, beta, depth)
		if score > alpha && score < beta {
			return score
		}
		if score >= beta {
			beta = valueInfinity
		}
		if score <= alpha {
			alpha = -valueInfinity
		}
		score = searchRoot(t, ml, alpha, beta, depth)
		if score > alpha && score < beta {
			return score
		}
	}
	return searchRoot(t, ml, -valueInfinity, valueInfinity, depth)
}

func searchRoot(t *thread, ml []Move, alpha, beta, depth int) int {
	const height = 0
	t.stack[height].pv.clear()
	var p = &t.stack[height].position
	t.evaluator.Init(p)
	t.stack[height].staticEval = t.evaluator.EvaluateQuick(p)
	var child = &t.stack[height+1].position
	var bestMoveIndex = 0
	for i, move := range ml {
		p.MakeMove(move, child)
		t.evaluator.MakeMove(p, move)
		var extension, reduction int
		extension = t.extend(depth, height)
		if depth >= 3 && i > 0 &&
			!(isCaptureOrPromotion(move)) {
			reduction = t.engine.lateMoveReduction(depth, i+1)
			reduction -= 2
			if p.IsCheck() || child.IsCheck() {
				reduction--
			}
			reduction = Max(reduction, 0) + extension
			reduction = Max(0, Min(depth-2, reduction))
		}
		var newDepth = depth - 1 + extension
		var score = alpha + 1
		// LMR
		if reduction > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth-reduction, height+1)
		}
		// PVS
		if score > alpha && beta != alpha+1 && i > 0 && newDepth > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth, height+1)
		}
		// full search
		if score > alpha {
			score = -t.alphaBeta(-beta, -alpha, newDepth, height+1)
		}
		t.evaluator.UnmakeMove()
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
func (t *thread) alphaBeta(alpha, beta, depth, height int) int {
	var oldAlpha = alpha
	var pvNode = beta != oldAlpha+1
	var newDepth int
	t.stack[height].pv.clear()

	var position = &t.stack[height].position

	if height >= maxHeight {
		return t.evaluator.EvaluateQuick(position)
	}

	if t.isRepeat(height) {
		return valueDraw
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
		if ttDepth >= depth && !pvNode && !(position.LastMove == MoveEmpty) {
			if ttValue >= beta && (ttBound&boundLower) != 0 {
				if ttMove != MoveEmpty && !isCaptureOrPromotion(ttMove) {
					t.updateKiller(ttMove, height)
				}
				return ttValue
			}
			if ttValue <= alpha && (ttBound&boundUpper) != 0 {
				return ttValue
			}
		}
	}

	var staticEval = t.evaluator.EvaluateQuick(position)
	t.stack[height].staticEval = staticEval
	var improving = height < 2 || staticEval > t.stack[height-2].staticEval

	// reverse futility pruning
	if !pvNode && depth <= 8 && !isCheck {
		var score = staticEval - pawnValue*depth
		if score >= beta {
			return staticEval
		}
	}

	if height+2 <= maxHeight {
		t.stack[height+2].killer1 = MoveEmpty
		t.stack[height+2].killer2 = MoveEmpty
	}

	// null-move pruning
	var child = &t.stack[height+1].position
	if !pvNode && depth >= 2 && !isCheck &&
		position.LastMove != MoveEmpty &&
		(height <= 1 || t.stack[height-1].position.LastMove != MoveEmpty) &&
		beta < valueWin &&
		!(ttHit && ttValue < beta && (ttBound&boundUpper) != 0) &&
		!isLateEndgame(position, position.WhiteMove) &&
		staticEval >= beta {
		var reduction = 4 + depth/6 + Min(2, (staticEval-beta)/200)
		position.MakeNullMove(child)
		t.evaluator.MakeMove(position, MoveEmpty)
		var score = -t.alphaBeta(-beta, -(beta - 1), depth-reduction, height+1)
		t.evaluator.UnmakeMove()
		if score >= beta {
			if score >= valueWin {
				score = beta
			}
			return score
		}
	}

	if !pvNode && depth >= 5 && !isCheck &&
		beta > valueLoss && beta < valueWin &&
		!(ttHit && ttDepth >= depth-4 && ttValue < beta && (ttBound&boundUpper) != 0) {

		var probcutBeta = Min(valueWin-1, beta+150)

		var mi = moveIteratorQS{
			position: position,
			buffer:   t.stack[height].moveList[:],
		}
		mi.Init()

		for mi.Reset(); ; {
			var move = mi.Next()
			if move == MoveEmpty {
				break
			}
			if !seeGEZero(position, move) {
				continue
			}
			if !position.MakeMove(move, child) {
				continue
			}
			t.evaluator.MakeMove(position, move)
			var score = -t.quiescence(-probcutBeta, -probcutBeta+1, height+1)
			if score >= probcutBeta {
				score = -t.alphaBeta(-probcutBeta, -probcutBeta+1, depth-4, height+1)
			}
			t.evaluator.UnmakeMove()
			if score >= probcutBeta {
				return score
			}
		}
	}

	// Internal iterative deepening
	if pvNode && depth >= 8 && ttMove == MoveEmpty {
		var iidDepth = depth - depth/4 - 5
		t.alphaBeta(alpha, beta, iidDepth, height)
		if t.stack[height].pv.size != 0 {
			ttMove = t.stack[height].pv.items[0]
			t.stack[height].pv.clear()
		}
	}

	var followUp Move
	if height > 0 {
		followUp = t.stack[height-1].position.LastMove
	}
	var historyContext = t.history.getContext(position.WhiteMove, position.LastMove, followUp)

	var mi = moveIterator{
		position:  position,
		buffer:    t.stack[height].moveList[:],
		history:   historyContext,
		transMove: ttMove,
		killer1:   t.stack[height].killer1,
		killer2:   t.stack[height].killer2,
	}
	mi.Init()

	// singular extension
	var ttMoveIsSingular = false
	if depth >= 8 &&
		ttHit && ttMove != MoveEmpty &&
		(ttBound&boundLower) != 0 && ttDepth >= depth-3 &&
		ttValue > valueLoss && ttValue < valueWin {

		ttMoveIsSingular = true
		var singularBeta = Max(-valueInfinity, ttValue-depth)
		newDepth = depth/2 - 1
		var quietsPlayed = 0
		for mi.Reset(); ; {
			var move = mi.Next()
			if move == MoveEmpty {
				break
			}
			if quietsPlayed >= 6 && !isCaptureOrPromotion(move) {
				continue
			}
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
			t.evaluator.MakeMove(position, move)
			if !isCaptureOrPromotion(move) {
				quietsPlayed++
			}
			var score = -t.alphaBeta(-singularBeta, -singularBeta+1, newDepth, height+1)
			t.evaluator.UnmakeMove()
			if score >= singularBeta {
				ttMoveIsSingular = false
				break
			}
		}
	}

	var movesSearched = 0
	var hasLegalMove = false
	var quietsSeen = 0

	var quietsSearched = t.stack[height].quietsSearched[:0]
	var bestMove Move

	var lmp = 5 + (depth-1)*depth
	if !improving {
		lmp /= 2
	}

	var best = -valueInfinity

	for mi.Reset(); ; {
		var move = mi.Next()
		if move == MoveEmpty {
			break
		}
		var isNoisy = isCaptureOrPromotion(move)
		if !isNoisy {
			quietsSeen++
		}

		if depth <= 8 && best > valueLoss && hasLegalMove && !isCheck {
			// late-move pruning
			if !(isNoisy ||
				move == mi.killer1 ||
				move == mi.killer2) &&
				quietsSeen > lmp {
				continue
			}

			// futility pruning
			if !(isNoisy ||
				move == mi.killer1 ||
				move == mi.killer2) &&
				staticEval+100+pawnValue*depth <= alpha {
				continue
			}

			// SEE pruning
			var seeMargin int
			if isNoisy {
				seeMargin = Max(depth, (staticEval+pawnValue-alpha)/pawnValue)
			} else {
				seeMargin = depth / 2
			}
			if !SeeGE(position, move, -seeMargin) {
				continue
			}
		}

		if !position.MakeMove(move, child) {
			continue
		}
		t.evaluator.MakeMove(position, move)
		hasLegalMove = true

		movesSearched++

		var extension, reduction int

		extension = t.extend(depth, height)
		if move == ttMove && ttMoveIsSingular {
			extension = 1
		}

		if depth >= 3 && movesSearched > 1 &&
			!(isNoisy) {
			reduction = t.engine.lateMoveReduction(depth, movesSearched)
			if move == mi.killer1 || move == mi.killer2 {
				reduction--
			}
			if !isCheck {
				var history = historyContext.ReadTotal(position.WhiteMove, move)
				reduction -= Max(-2, Min(2, history/5000))

				if !improving {
					reduction++
				}
			}
			if pvNode {
				reduction -= 2
			}
			if isCheck || child.IsCheck() {
				reduction--
			}
			reduction = Max(reduction, 0) + extension
			reduction = Max(0, Min(depth-2, reduction))
		}

		if !isNoisy {
			quietsSearched = append(quietsSearched, move)
		}

		newDepth = depth - 1 + extension

		var score = alpha + 1
		// LMR
		if reduction > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth-reduction, height+1)
		}
		// PVS
		if score > alpha && beta != alpha+1 && movesSearched > 1 && newDepth > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth, height+1)
		}
		// full search
		if score > alpha {
			score = -t.alphaBeta(-beta, -alpha, newDepth, height+1)
		}

		t.evaluator.UnmakeMove()

		if score > best {
			best = score
			bestMove = move
		}
		if score > alpha {
			alpha = score
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

	if alpha > oldAlpha && bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		historyContext.Update(position.WhiteMove, quietsSearched, bestMove, depth)
		t.updateKiller(bestMove, height)
	}

	ttBound = 0
	if best > oldAlpha {
		ttBound |= boundLower
	}
	if best < beta {
		ttBound |= boundUpper
	}
	t.engine.transTable.Update(position.Key, depth, valueToTT(best, height), ttBound, bestMove)

	return best
}

func (t *thread) quiescence(alpha, beta, height int) int {
	t.stack[height].pv.clear()
	t.incNodes()
	var position = &t.stack[height].position
	if isDraw(position) {
		return valueDraw
	}
	if height >= maxHeight {
		return t.evaluator.EvaluateQuick(position)
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
	var best = -valueInfinity
	if !isCheck {
		var eval = t.evaluator.EvaluateQuick(position)
		best = Max(best, eval)
		if eval > alpha {
			alpha = eval
			if alpha >= beta {
				return alpha
			}
		}
	}
	var mi = moveIteratorQS{
		position: position,
		buffer:   t.stack[height].moveList[:],
	}
	mi.Init()
	var hasLegalMove = false
	var child = &t.stack[height+1].position
	for mi.Reset(); ; {
		var move = mi.Next()
		if move == MoveEmpty {
			break
		}
		if !isCheck && !seeGEZero(position, move) {
			continue
		}
		if !position.MakeMove(move, child) {
			continue
		}
		t.evaluator.MakeMove(position, move)
		hasLegalMove = true
		var score = -t.quiescence(-beta, -alpha, height+1)
		t.evaluator.UnmakeMove()
		best = Max(best, score)
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
	return best
}

func (t *thread) incNodes() {
	t.nodes++
	if t.nodes&255 == 0 {
		//fixed nodes search only in single threaded mode
		if t.engine.Threads == 1 {
			t.engine.timeManager.OnNodesChanged(int(t.engine.nodes + t.nodes))
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

func (t *thread) isRepeat(height int) bool {
	var p = &t.stack[height].position

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

	return t.engine.historyKeys[p.Key] >= 2
}

func (t *thread) extend(depth, height int) int {
	//var p = &t.stack[height].position
	var child = &t.stack[height+1].position
	var givesCheck = child.IsCheck()

	if givesCheck && depth >= 3 {
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

	var historyContext = t.history.getContext(p.WhiteMove, p.LastMove, MoveEmpty)
	var mi = moveIterator{
		position:  p,
		buffer:    t.stack[height].moveList[:],
		history:   historyContext,
		transMove: transMove,
	}
	mi.Init()

	var result []Move
	var child = &t.stack[height+1].position
	for mi.Reset(); ; {
		var move = mi.Next()
		if move == MoveEmpty {
			break
		}
		if p.MakeMove(move, child) {
			result = append(result, move)
		}
	}
	return result
}

func (t *thread) updateKiller(move Move, height int) {
	if t.stack[height].killer1 != move {
		t.stack[height].killer2 = t.stack[height].killer1
		t.stack[height].killer1 = move
	}
}
