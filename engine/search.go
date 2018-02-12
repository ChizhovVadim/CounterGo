package engine

import (
	"sync"

	. "github.com/ChizhovVadim/CounterGo/common"
)

func recoverFromSearchTimeout() {
	var r = recover()
	if r != nil && r != searchTimeout {
		panic(r)
	}
}

func (node *node) IterateSearch(progress func(SearchInfo)) (result SearchInfo) {
	defer recoverFromSearchTimeout()
	var engine = node.engine
	defer func() {
		result.Time = engine.timeManager.ElapsedMilliseconds()
		result.Nodes = engine.timeManager.Nodes()
	}()

	var p = node.position
	var ml = p.GenerateLegalMoves()
	if len(ml) == 0 {
		return
	}
	result.MainLine = []Move{ml[0]}
	if len(ml) == 1 {
		return
	}
	node.sortRootMoves(ml)
	const height = 0
	const beta = valueInfinity
	var gate sync.Mutex
	var prevScore int
	for depth := 2; depth <= maxHeight; depth++ {
		var alpha = -valueInfinity
		var bestMoveIndex = 0
		{
			var child = node.next()
			var move = ml[0]
			p.MakeMove(move, child.position)
			var newDepth = node.newDepth(depth, child)
			var score = -child.alphaBeta(-beta, -alpha, newDepth)
			alpha = score
			result = SearchInfo{
				Depth:    depth,
				Score:    newUciScore(score),
				MainLine: append([]Move{move}, child.principalVariation...),
				Time:     engine.timeManager.ElapsedMilliseconds(),
				Nodes:    engine.timeManager.Nodes(),
			}
			if progress != nil {
				progress(result)
			}
		}
		engine.initKillers()
		var index = 1
		parallelDo(engine.Threads.Value, func(threadIndex int) {
			defer recoverFromSearchTimeout()
			var child = node.nextOnThread(threadIndex)
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
				p.MakeMove(move, child.position)
				var newDepth = node.newDepth(depth, child)
				var score = -child.alphaBeta(-(localAlpha + 1), -localAlpha, newDepth)
				if score > localAlpha {
					score = -child.alphaBeta(-beta, -localAlpha, newDepth)
				}
				gate.Lock()
				if score > alpha {
					alpha = score
					result = SearchInfo{
						Depth:    depth,
						Score:    newUciScore(score),
						MainLine: append([]Move{move}, child.principalVariation...),
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
		if engine.timeManager.IsHardTimeout() {
			break
		}
		if alpha >= winIn(depth) || alpha <= lossIn(depth) {
			break
		}
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && engine.timeManager.IsSoftTimeout() {
			break
		}
		moveToBegin(ml, bestMoveIndex)
		hashStorePV(node, result.Depth, alpha, result.MainLine)
		prevScore = alpha
	}
	return
}

func (node *node) sortRootMoves(moves []Move) {
	var child = node.next()
	var list = make([]orderedMove, len(moves))
	for i, m := range moves {
		node.position.MakeMove(m, child.position)
		var score = -child.quiescence(-valueInfinity, valueInfinity, 1)
		list[i] = orderedMove{m, score}
	}
	sortMoves(list)
	for i := range moves {
		moves[i] = list[i].move
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

func hashStorePV(node *node, depth, score int, pv []Move) {
	for _, move := range pv {
		node.engine.transTable.Update(node.position, depth,
			valueToTT(score, node.height), boundLower|boundUpper, move)
		var child = node.next()
		node.position.MakeMove(move, child.position)
		depth = node.newDepth(depth, child)
		if depth <= 0 {
			break
		}
		node = child
	}
}

func (node *node) alphaBeta(alpha, beta, depth int) int {
	var newDepth, score int
	node.clearPV()

	if node.height >= maxHeight || node.isDraw() {
		return valueDraw
	}

	if depth <= 0 {
		return node.quiescence(alpha, beta, 1)
	}

	var engine = node.engine
	engine.timeManager.IncNodes()

	var position = node.position
	var isCheck = position.IsCheck()

	if winIn(node.height+1) <= alpha {
		return alpha
	}
	if lossIn(node.height+2) >= beta && !isCheck {
		return beta
	}

	var hashMove = MoveEmpty

	if ttDepth, ttScore, ttType, ttMove, ok := engine.transTable.Read(position); ok {
		hashMove = ttMove
		if ttDepth >= depth {
			ttScore = valueFromTT(ttScore, node.height)
			if ttScore >= beta && (ttType&boundLower) != 0 {
				return beta
			}
			if ttScore <= alpha && (ttType&boundUpper) != 0 {
				return alpha
			}
		}
	}

	var child = node.next()
	if depth >= 2 && !isCheck && position.LastMove != MoveEmpty &&
		beta < valueWin &&
		!isLateEndgame(position, position.WhiteMove) {
		newDepth = depth - 4
		position.MakeNullMove(child.position)
		var rbeta = beta + 15 // STM-bonus
		if newDepth <= 0 {
			score = -child.quiescence(-rbeta, -(rbeta - 1), 1)
		} else {
			score = -child.alphaBeta(-rbeta, -(rbeta - 1), newDepth)
		}
		if score >= rbeta && score < valueWin {
			return beta
		}
	}

	if depth >= 4 && hashMove == MoveEmpty &&
		beta > alpha+PawnValue/2 {
		//good test: position fen 8/pp6/2p5/P1P5/1P3k2/3K4/8/8 w - - 5 47
		node.alphaBeta(alpha, beta, depth-2)
		hashMove = node.bestMove()
		node.clearPV() //!
	}

	var moveSort = moveSort{
		node:  node,
		trans: hashMove,
	}

	var moveCount = 0
	node.quietsSearched = node.quietsSearched[:0]
	var staticEval = valueInfinity

	for {
		var move = moveSort.Next()
		if move == MoveEmpty {
			break
		}

		if position.MakeMove(move, child.position) {
			moveCount++

			newDepth = node.newDepth(depth, child)
			var reduction = 0

			if !isCaptureOrPromotion(move) && moveCount > 1 &&
				!isCheck && !child.position.IsCheck() &&
				move != node.killer1 && move != node.killer2 &&
				!isPawnPush7th(move, position.WhiteMove) &&
				alpha > valueLoss {

				if depth <= 2 {
					if staticEval == valueInfinity {
						staticEval = engine.evaluate(position)
					}
					if staticEval+PawnValue <= alpha {
						continue
					}
				}

				if depth >= 3 && !isPawnAdvance(move, position.WhiteMove) {
					reduction = engine.lateMoveReduction(depth, moveCount)
				}
			}

			if !isCaptureOrPromotion(move) {
				node.quietsSearched = append(node.quietsSearched, move)
			}

			if reduction > 0 {
				score = -child.alphaBeta(-(alpha + 1), -alpha, depth-1-reduction)
				if score <= alpha {
					continue
				}
			}

			score = -child.alphaBeta(-beta, -alpha, newDepth)

			if score > alpha {
				alpha = score
				node.composePV(move, child)
				if alpha >= beta {
					break
				}
			}
		}
	}

	if moveCount == 0 {
		if isCheck {
			return lossIn(node.height)
		}
		return valueDraw
	}

	var bestMove = node.bestMove()

	if bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		engine.historyTable.Update(node, bestMove, depth)
	}

	var ttType = 0
	if bestMove != MoveEmpty {
		ttType |= boundLower
	}
	if alpha < beta {
		ttType |= boundUpper
	}
	engine.transTable.Update(position, depth, valueToTT(alpha, node.height), ttType, bestMove)

	return alpha
}

func (node *node) quiescence(alpha, beta, depth int) int {
	var engine = node.engine
	engine.timeManager.IncNodes()
	node.clearPV()
	if node.height >= maxHeight {
		return valueDraw
	}
	var position = node.position
	var isCheck = position.IsCheck()
	var eval = 0
	if !isCheck {
		eval = engine.evaluate(position)
		if eval > alpha {
			alpha = eval
		}
		if eval >= beta {
			return alpha
		}
	}
	var moveSort = moveSortQS{
		node:      node,
		genChecks: depth > 0,
	}
	var moveCount = 0
	var child = node.next()
	for {
		var move = moveSort.Next()
		if move == MoveEmpty {
			break
		}
		var danger = isDangerCapture(position, move)
		if !isCheck && !danger && !seeGEZero(position, move) {
			continue
		}
		if position.MakeMove(move, child.position) {
			moveCount++
			if !isCheck && !danger && !child.position.IsCheck() &&
				eval+moveValue(move)+PawnValue <= alpha {
				continue
			}
			var score = -child.quiescence(-beta, -alpha, depth-1)
			if score > alpha {
				alpha = score
				node.composePV(move, child)
				if score >= beta {
					break
				}
			}
		}
	}
	if isCheck && moveCount == 0 {
		return lossIn(node.height)
	}
	return alpha
}

func (node *node) newDepth(depth int, child *node) int {
	var p = node.position
	var prevMove = p.LastMove
	var move = child.position.LastMove
	var givesCheck = child.position.IsCheck()

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
