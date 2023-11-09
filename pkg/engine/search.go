package engine

import (
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const pawnValue = 100

func aspirationWindow(t *thread, ml []Move, depth, prevScore int) int {
	t.rootDepth = depth
	if t.engine.Options.AspirationWindows &&
		depth >= 5 && !(prevScore <= valueLoss || prevScore >= valueWin) {
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
	var p = &t.stack[height].position
	t.evaluator.Init(p)
	return t.alphaBeta(alpha, beta, depth, height, 0)
}

// main search method
func (t *thread) alphaBeta(alpha, beta, depth, height int, skipMove Move) int {
	if depth <= 0 {
		return t.quiescence(alpha, beta, height)
	}
	t.clearPV(height)

	var rootNode = height == 0
	var pvNode = beta != alpha+1
	var position = &t.stack[height].position
	var isCheck = position.IsCheck()
	var ttMoveIsSingular = false

	if !rootNode {
		if height >= maxHeight {
			return t.evaluator.EvaluateQuick(position)
		}
		if t.isRepeat(height) {
			return valueDraw
		}
		if isDraw(position) {
			return valueDraw
		}
		// mate distance pruning
		if winIn(height+1) <= alpha {
			return alpha
		}
		if lossIn(height+2) >= beta && !isCheck {
			return beta
		}
	}

	// transposition table
	var (
		ttDepth, ttValue, ttBound int
		ttMove                    Move
		ttHit                     bool
	)
	if skipMove == 0 {
		ttDepth, ttValue, ttBound, ttMove, ttHit = t.engine.transTable.Read(position.Key)
	}
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

	var options = &t.engine.Options
	if height+2 <= maxHeight {
		t.stack[height+2].killer1 = MoveEmpty
		t.stack[height+2].killer2 = MoveEmpty
	}
	var child = &t.stack[height+1].position

	if !rootNode && skipMove == 0 {

		// reverse futility pruning
		if options.ReverseFutility && !pvNode && depth <= 8 && !isCheck {
			var score = staticEval - pawnValue*depth
			if score >= beta {
				return staticEval
			}
		}

		// null-move pruning
		if options.NullMovePruning && !pvNode && depth >= 2 && !isCheck &&
			position.LastMove != MoveEmpty &&
			(height <= 1 || t.stack[height-1].position.LastMove != MoveEmpty) &&
			beta < valueWin &&
			!(ttHit && ttValue < beta && (ttBound&boundUpper) != 0) &&
			!isLateEndgame(position, position.WhiteMove) &&
			staticEval >= beta {
			var reduction = 4 + depth/6 + Min(2, (staticEval-beta)/200)
			t.MakeMove(MoveEmpty, height)
			var score = -t.alphaBeta(-beta, -(beta - 1), depth-reduction, height+1, 0)
			t.UnmakeMove()
			if score >= beta {
				if score >= valueWin {
					score = beta
				}
				return score
			}
		}

		var probcutBeta = Min(valueWin-1, beta+150)
		if options.Probcut && !pvNode && depth >= 5 && !isCheck &&
			beta > valueLoss && beta < valueWin &&
			!(ttHit && ttDepth >= depth-4 && ttValue < probcutBeta && (ttBound&boundUpper) != 0) {

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
				if !t.MakeMove(move, height) {
					continue
				}
				var score = -t.quiescence(-probcutBeta, -probcutBeta+1, height+1)
				if score >= probcutBeta {
					score = -t.alphaBeta(-probcutBeta, -probcutBeta+1, depth-4, height+1, 0)
				}
				t.UnmakeMove()
				if score >= probcutBeta {
					return score
				}
			}
		}

		// singular extension
		if options.SingularExt && depth >= 8 &&
			ttHit && ttMove != MoveEmpty &&
			(ttBound&boundLower) != 0 && ttDepth >= depth-3 &&
			ttValue > valueLoss && ttValue < valueWin {
			var singularBeta = Max(-valueInfinity, ttValue-depth)
			var score = t.alphaBeta(singularBeta-1, singularBeta, depth/2, height, ttMove)
			ttMoveIsSingular = score < singularBeta
		}
	}

	var historyContext = t.getHistoryContext(height)

	var mi = t.initMoveIterator(height, ttMove)
	var killer1 = t.stack[height].killer1
	var killer2 = t.stack[height].killer2

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
	var oldAlpha = alpha

	for mi.Reset(); ; {
		var move = mi.Next()
		if move == MoveEmpty {
			break
		}
		if move == skipMove {
			continue
		}
		var isNoisy = isCaptureOrPromotion(move)
		if !isNoisy {
			quietsSeen++
		}

		if depth <= 8 && best > valueLoss && hasLegalMove && !isCheck && !rootNode {
			// late-move pruning
			if options.Lmp && !(isNoisy ||
				move == killer1 ||
				move == killer2) &&
				quietsSeen > lmp {
				continue
			}

			// futility pruning
			if options.Futility && !(isNoisy ||
				move == killer1 ||
				move == killer2) &&
				staticEval+100+pawnValue*depth <= alpha {
				continue
			}

			// SEE pruning
			if options.See {
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
		}

		if !t.MakeMove(move, height) {
			continue
		}
		hasLegalMove = true

		movesSearched++

		var extension, reduction int

		if options.CheckExt && child.IsCheck() && depth >= 3 {
			extension = 1
		}
		if move == ttMove && ttMoveIsSingular {
			extension = 1
		}

		if depth >= 3 && movesSearched > 1 &&
			!(isNoisy) {
			reduction = options.Lmr(depth, movesSearched)
			if move == killer1 || move == killer2 {
				reduction--
			}
			if !isCheck {
				var history = historyContext.ReadTotal(move)
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

		var newDepth = depth - 1 + extension

		var score = alpha + 1
		// LMR
		if reduction > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth-reduction, height+1, 0)
		}
		// PVS
		if score > alpha && beta != alpha+1 && movesSearched > 1 && newDepth > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth, height+1, 0)
		}
		// full search
		if score > alpha {
			score = -t.alphaBeta(-beta, -alpha, newDepth, height+1, 0)
		}

		t.UnmakeMove()

		if score > best {
			best = score
			bestMove = move
		}
		if score > alpha {
			alpha = score
			t.assignPV(height, move)
			if alpha >= beta {
				break
			}
		}
	}

	if !hasLegalMove {
		if !isCheck && skipMove == 0 {
			return valueDraw
		}
		return lossIn(height)
	}

	if alpha > oldAlpha && bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		historyContext.Update(quietsSearched, bestMove, depth)
		t.updateKiller(bestMove, height)
	}

	if skipMove == 0 {
		ttBound = 0
		if best > oldAlpha {
			ttBound |= boundLower
		}
		if best < beta {
			ttBound |= boundUpper
		}
		if !(rootNode && ttBound == boundUpper) {
			t.engine.transTable.Update(position.Key, depth, valueToTT(best, height), ttBound, bestMove)
		}
	}

	return best
}

func (t *thread) quiescence(alpha, beta, height int) int {
	t.clearPV(height)
	var position = &t.stack[height].position
	if isDraw(position) {
		return valueDraw
	}
	if height >= maxHeight {
		return t.evaluator.EvaluateQuick(position)
	}
	if t.isRepeat(height) {
		return valueDraw
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
	for mi.Reset(); ; {
		var move = mi.Next()
		if move == MoveEmpty {
			break
		}
		if !isCheck && !seeGEZero(position, move) {
			continue
		}
		if !t.MakeMove(move, height) {
			continue
		}
		hasLegalMove = true
		var score = -t.quiescence(-beta, -alpha, height+1)
		t.UnmakeMove()
		best = Max(best, score)
		if score > alpha {
			alpha = score
			t.assignPV(height, move)
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
		if t.engine.Options.Threads == 1 {
			t.engine.timeManager.OnNodesChanged(int(t.engine.mainLine.nodes + t.nodes))
		}
		if t.engine.timeManager.IsDone() {
			panic(errSearchTimeout)
		}
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

	var mi = t.initMoveIterator(height, transMove)

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

func (t *thread) MakeMove(move Move, height int) bool {
	var pos = &t.stack[height].position
	var child = &t.stack[height+1].position
	if move == MoveEmpty {
		pos.MakeNullMove(child)
	} else {
		if !pos.MakeMove(move, child) {
			return false
		}
	}
	t.evaluator.MakeMove(pos, move)
	t.incNodes()
	return true
}

func (t *thread) UnmakeMove() {
	t.evaluator.UnmakeMove()
}
