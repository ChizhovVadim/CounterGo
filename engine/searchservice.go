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

func (ctx *searchContext) IterateSearch(progress func(SearchInfo)) (result SearchInfo) {
	defer recoverFromSearchTimeout()
	var engine = ctx.engine
	defer func() {
		result.Time = engine.timeManager.ElapsedMilliseconds()
		result.Nodes = engine.timeManager.Nodes()
	}()

	var p = ctx.position
	var ml = GenerateLegalMoves(p)
	if len(ml) == 0 {
		return
	}
	result.MainLine = []Move{ml[0]}
	if len(ml) == 1 {
		return
	}
	ctx.sortRootMoves(ml)
	const height = 0
	const beta = ValueInfinity
	var gate sync.Mutex
	var prevScore int
	for depth := 2; depth <= MaxHeight; depth++ {
		var alpha = -ValueInfinity
		var bestMoveIndex = 0
		{
			var child = ctx.next()
			var move = ml[0]
			p.MakeMove(move, child.position)
			var newDepth = ctx.newDepth(depth, child)
			var score = -child.alphaBeta(-beta, -alpha, newDepth)
			alpha = score
			result = SearchInfo{
				Depth:    depth,
				Score:    scoreToUci(score),
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
			var child = ctx.nextOnThread(threadIndex)
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
				var newDepth = ctx.newDepth(depth, child)
				var score = -child.alphaBeta(-(localAlpha + 1), -localAlpha, newDepth)
				if score > localAlpha {
					score = -child.alphaBeta(-beta, -localAlpha, newDepth)
				}
				gate.Lock()
				if score > alpha {
					alpha = score
					result = SearchInfo{
						Depth:    depth,
						Score:    scoreToUci(score),
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
		if alpha >= mateIn(depth) || alpha <= matedIn(depth) {
			break
		}
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && engine.timeManager.IsSoftTimeout() {
			break
		}
		moveToBegin(ml, bestMoveIndex)
		hashStorePV(ctx, result.Depth, alpha, result.MainLine)
		prevScore = alpha
	}
	return
}

func (ctx *searchContext) sortRootMoves(moves []Move) {
	var child = ctx.next()
	var list = make([]orderedMove, len(moves))
	for i, m := range moves {
		ctx.position.MakeMove(m, child.position)
		var score = -child.quiescence(-ValueInfinity, ValueInfinity, 1)
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

func hashStorePV(ctx *searchContext, depth, score int, pv []Move) {
	for _, move := range pv {
		ctx.engine.transTable.Update(ctx.position, depth,
			valueToTT(score, ctx.height), boundLower|boundUpper, move)
		var child = ctx.next()
		ctx.position.MakeMove(move, child.position)
		depth = ctx.newDepth(depth, child)
		if depth <= 0 {
			break
		}
		ctx = child
	}
}

func (ctx *searchContext) alphaBeta(alpha, beta, depth int) int {
	var newDepth, score int
	ctx.clearPV()

	if ctx.height >= MaxHeight || ctx.isDraw() {
		return ValueDraw
	}

	if depth <= 0 {
		return ctx.quiescence(alpha, beta, 1)
	}

	var engine = ctx.engine
	engine.timeManager.IncNodes()

	if mateIn(ctx.height+1) <= alpha {
		return alpha
	}

	var position = ctx.position
	var hashMove = MoveEmpty

	if ttDepth, ttScore, ttType, ttMove, ok := engine.transTable.Read(position); ok {
		hashMove = ttMove
		if ttDepth >= depth {
			ttScore = valueFromTT(ttScore, ctx.height)
			if ttScore >= beta && (ttType&boundLower) != 0 {
				return beta
			}
			if ttScore <= alpha && (ttType&boundUpper) != 0 {
				return alpha
			}
		}
	}

	var isCheck = position.IsCheck()

	var child = ctx.next()
	if depth >= 2 && !isCheck && position.LastMove != MoveEmpty &&
		beta < ValueMateInMaxHeight &&
		!isLateEndgame(position, position.WhiteMove) {
		newDepth = depth - 4
		position.MakeNullMove(child.position)
		if newDepth <= 0 {
			score = -child.quiescence(-beta, -(beta - 1), 1)
		} else {
			score = -child.alphaBeta(-beta, -(beta - 1), newDepth)
		}
		if score >= beta && score < ValueMateInMaxHeight {
			return beta
		}
	}

	if depth >= 4 && hashMove == MoveEmpty && beta > alpha+PawnValue/2 {
		//good test: position fen 8/pp6/2p5/P1P5/1P3k2/3K4/8/8 w - - 5 47
		ctx.alphaBeta(alpha, beta, depth-2)
		hashMove = ctx.bestMove()
		ctx.clearPV() //!
	}

	var moveSort = moveSort{
		ctx:   ctx,
		trans: hashMove,
	}

	var moveCount = 0
	ctx.quietsSearched = ctx.quietsSearched[:0]
	var staticEval = ValueInfinity

	for {
		var move = moveSort.Next()
		if move == MoveEmpty {
			break
		}

		if position.MakeMove(move, child.position) {
			moveCount++

			newDepth = ctx.newDepth(depth, child)
			var reduction = 0

			if !isCaptureOrPromotion(move) && moveCount > 1 &&
				!isCheck && !child.position.IsCheck() &&
				move != ctx.killer1 && move != ctx.killer2 &&
				!isPawnPush7th(move, position.WhiteMove) &&
				alpha > ValueMatedInMaxHeight {

				if depth <= 2 {
					if staticEval == ValueInfinity {
						staticEval = engine.evaluate(position)
					}
					if staticEval+PawnValue <= alpha {
						continue
					}
				}

				if depth >= 3 && !isPawnAdvance(move, position.WhiteMove) {
					reduction = 1
				}
			}

			if !isCaptureOrPromotion(move) {
				ctx.quietsSearched = append(ctx.quietsSearched, move)
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
				ctx.composePV(move, child)
				if alpha >= beta {
					break
				}
			}
		}
	}

	if moveCount == 0 {
		if isCheck {
			return matedIn(ctx.height)
		}
		return ValueDraw
	}

	var bestMove = ctx.bestMove()

	if bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		engine.historyTable.Update(ctx, bestMove, depth)
	}

	var ttType = 0
	if bestMove != MoveEmpty {
		ttType |= boundLower
	}
	if alpha < beta {
		ttType |= boundUpper
	}
	engine.transTable.Update(position, depth, valueToTT(alpha, ctx.height), ttType, bestMove)

	return alpha
}

func (ctx *searchContext) quiescence(alpha, beta, depth int) int {
	var engine = ctx.engine
	engine.timeManager.IncNodes()
	ctx.clearPV()
	if ctx.height >= MaxHeight {
		return ValueDraw
	}
	var position = ctx.position
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
		ctx:       ctx,
		genChecks: depth > 0,
	}
	var moveCount = 0
	var child = ctx.next()
	for {
		var move = moveSort.Next()
		if move == MoveEmpty {
			break
		}
		var danger = isDangerCapture(position, move)
		if !isCheck && !danger && !SEE_GE(position, move) {
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
				ctx.composePV(move, child)
				if score >= beta {
					break
				}
			}
		}
	}
	if isCheck && moveCount == 0 {
		return matedIn(ctx.height)
	}
	return alpha
}

func (ctx *searchContext) newDepth(depth int, child *searchContext) int {
	var p = ctx.position
	var prevMove = p.LastMove
	var move = child.position.LastMove
	var givesCheck = child.position.IsCheck()

	if prevMove != MoveEmpty &&
		prevMove.To() == move.To() &&
		move.CapturedPiece() > Pawn &&
		prevMove.CapturedPiece() > Pawn &&
		SEE_GE(p, move) {
		return depth
	}

	if givesCheck && (depth <= 1 || SEE_GE(p, move)) {
		return depth
	}

	if isPawnPush7th(move, p.WhiteMove) && SEE_GE(p, move) {
		return depth
	}

	return depth - 1
}
