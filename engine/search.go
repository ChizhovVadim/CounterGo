package engine

import (
	"sync"

	. "github.com/ChizhovVadim/CounterGo/common"
)

func (engine *Engine) iterateSearch(progress func(SearchInfo)) (result SearchInfo) {
	defer recoverFromSearchTimeout()
	var mainThread = &engine.threads[0]
	defer func() {
		result.Time = engine.timeManager.ElapsedMilliseconds()
		result.Nodes = engine.timeManager.Nodes()
	}()

	const height = 0
	var p = &mainThread.stack[height].position
	var ml = p.GenerateLegalMoves()
	if len(ml) == 0 {
		return
	}
	result.MainLine = []Move{ml[0]}
	if len(ml) == 1 {
		return
	}
	mainThread.sortRootMoves(ml)
	const beta = valueInfinity
	var gate sync.Mutex
	var prevScore int
	for depth := 2; depth <= maxHeight; depth++ {
		var alpha = -valueInfinity
		var bestMoveIndex = 0
		{
			var child = &mainThread.stack[height+1].position
			var move = ml[0]
			p.MakeMove(move, child)
			var newDepth = mainThread.newDepth(depth, height)
			var score = -mainThread.alphaBeta(-beta, -alpha, newDepth, height+1)
			engine.timeManager.nodes += int64(mainThread.nodes)
			mainThread.nodes = 0
			alpha = score
			result = SearchInfo{
				Depth:    depth,
				Score:    newUciScore(score),
				MainLine: append([]Move{move}, mainThread.pvTable.Read(child)...),
				Time:     engine.timeManager.ElapsedMilliseconds(),
				Nodes:    engine.timeManager.Nodes(),
			}
			if progress != nil {
				progress(result)
			}
		}
		var index = 1
		parallelDo(engine.Threads.Value, func(threadIndex int) {
			defer recoverFromSearchTimeout()
			var thread = &engine.threads[threadIndex]
			var child = &thread.stack[height+1].position
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
				var newDepth = thread.newDepth(depth, height)
				var score = -thread.alphaBeta(-(localAlpha + 1), -localAlpha, newDepth, height+1)
				if score > localAlpha {
					score = -thread.alphaBeta(-beta, -localAlpha, newDepth, height+1)
				}
				gate.Lock()
				engine.timeManager.nodes += int64(thread.nodes)
				thread.nodes = 0
				if score > alpha {
					alpha = score
					result = SearchInfo{
						Depth:    depth,
						Score:    newUciScore(score),
						MainLine: append([]Move{move}, thread.pvTable.Read(child)...),
						Time:     engine.timeManager.ElapsedMilliseconds(),
						Nodes:    engine.timeManager.Nodes(),
					}
					if progress != nil {
						progress(result)
					}
					bestMoveIndex = localIndex
				}
				gate.Unlock()
			}
		})
		if isDone(mainThread.done) {
			break
		}
		if alpha >= winIn(depth-3) || alpha <= lossIn(depth-3) {
			break
		}
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && engine.timeManager.IsSoftTimeout() {
			break
		}
		moveToBegin(ml, bestMoveIndex)
		//TODO hashStorePV(node, result.Depth, alpha, result.MainLine)
		prevScore = alpha
	}
	return
}

func (t *thread) alphaBeta(alpha, beta, depth, height int) int {
	var newDepth, score int

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

	if ttDepth, ttScore, ttType, ttMove, ok := t.engine.transTable.Read(position); ok {
		hashMove = ttMove
		if ttDepth >= depth {
			ttScore = valueFromTT(ttScore, height)
			if ttScore >= beta && (ttType&boundLower) != 0 {
				return beta
			}
			if ttScore <= alpha && (ttType&boundUpper) != 0 {
				return alpha
			}
		}
	}

	var child = &t.stack[height+1].position
	if depth >= 2 && !isCheck && position.LastMove != MoveEmpty &&
		beta < valueWin &&
		!isLateEndgame(position, position.WhiteMove) {
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

	if depth >= 4 && hashMove == MoveEmpty &&
		beta > alpha+PawnValue/2 {
		//good test: position fen 8/pp6/2p5/P1P5/1P3k2/3K4/8/8 w - - 5 47
		t.alphaBeta(alpha, beta, depth-2, height)
		_, _, _, hashMove, _ = t.engine.transTable.Read(position)
	}

	var ml = position.GenerateMoves(t.stack[height].moveList[:])
	t.sortTable.Note(position, ml, hashMove, height)

	var moveCount = 0
	var quietsSearched = t.stack[height].quietsSearched[:0]
	var staticEval = valueInfinity
	var bestMove Move
	const DirectCount = 4

	for i := range ml {
		if i < DirectCount {
			moveToTop(ml[i:])
		} else if i == DirectCount {
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
				!isPawnPush7th(move, position.WhiteMove) &&
				alpha > valueLoss {

				if depth <= 2 {
					if staticEval == valueInfinity {
						staticEval = t.evaluator.Evaluate(position)
					}
					if staticEval+PawnValue*depth <= alpha {
						continue
					}
				}

				if depth >= 3 && !isPawnAdvance(move, position.WhiteMove) {
					reduction = t.engine.lateMoveReduction(depth, moveCount)
				}
			}

			if !isCaptureOrPromotion(move) {
				quietsSearched = append(quietsSearched, move)
			}

			if reduction > 0 {
				score = -t.alphaBeta(-(alpha + 1), -alpha, depth-1-reduction, height+1)
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
	if ttType == boundExact {
		t.pvTable.Save(position, bestMove)
	}
	t.engine.transTable.Update(position, depth, valueToTT(alpha, height), ttType, bestMove)

	return alpha
}

func (t *thread) quiescence(alpha, beta, depth, height int) int {
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
	var bestMove = MoveEmpty
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
				bestMove = move
				if score >= beta {
					break
				}
			}
		}
	}
	if isCheck && moveCount == 0 {
		return lossIn(height)
	}
	if bestMove != MoveEmpty && alpha < beta {
		t.pvTable.Save(position, bestMove)
	}
	return alpha
}

func (t *thread) incNodes() {
	t.nodes++
	if (t.nodes&255) == 0 && isDone(t.done) {
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

func (t *thread) sortRootMoves(moves []Move) {
	const height = 0
	var position = &t.stack[height].position
	var child = &t.stack[height+1].position
	var ml = t.stack[height].moveList[:0]
	for _, m := range moves {
		position.MakeMove(m, child)
		var score = -t.quiescence(-valueInfinity, valueInfinity, 1, height+1)
		ml = append(ml, OrderedMove{Move: m, Key: score})
	}
	sortMoves(ml)
	for i := range moves {
		moves[i] = ml[i].Move
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

/*func hashStorePV(node *node, depth, score int, pv []Move) {
	for _, move := range pv {
		node.engine.transTable.Update(&node.position, depth,
			valueToTT(score, node.height), boundLower|boundUpper, move)
		var child = node.next()
		node.position.MakeMove(move, &child.position)
		depth = node.newDepth(depth, child)
		if depth <= 0 {
			break
		}
		node = child
	}
}*/

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
