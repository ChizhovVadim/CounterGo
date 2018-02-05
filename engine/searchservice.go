package engine

import (
	"sync"

	. "github.com/ChizhovVadim/CounterGo/common"
)

func RecoverFromSearchTimeout() {
	var r = recover()
	if r != nil && r != searchTimeout {
		panic(r)
	}
}

func (ctx *searchContext) IterateSearch(progress func(SearchInfo)) (result SearchInfo) {
	defer RecoverFromSearchTimeout()
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
	const beta = VALUE_INFINITE
	var gate sync.Mutex
	for depth := 2; depth <= MAX_HEIGHT; depth++ {
		var prevScore = result.Score
		var alpha = -VALUE_INFINITE
		var bestMoveIndex = 0
		{
			var child = ctx.Next()
			var move = ml[0]
			p.MakeMove(move, child.position)
			var newDepth = ctx.NewDepth(depth, child)
			var score = -child.AlphaBeta(-beta, -alpha, newDepth)
			alpha = score
			result = SearchInfo{
				Depth:    depth,
				Score:    score,
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
		ParallelDo(engine.Threads.Value, func(threadIndex int) {
			defer RecoverFromSearchTimeout()
			var child = ctx.NextOnThread(threadIndex)
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
				var newDepth = ctx.NewDepth(depth, child)
				var score = -child.AlphaBeta(-(localAlpha + 1), -localAlpha, newDepth)
				if score > localAlpha {
					score = -child.AlphaBeta(-beta, -localAlpha, newDepth)
				}
				gate.Lock()
				if score > alpha {
					alpha = score
					result = SearchInfo{
						Depth:    depth,
						Score:    score,
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
		if alpha >= MateIn(depth) || alpha <= MatedIn(depth) {
			break
		}
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && engine.timeManager.IsSoftTimeout() {
			break
		}
		moveToBegin(ml, bestMoveIndex)
		HashStorePV(ctx, result.Depth, result.Score, result.MainLine)
	}
	return
}

func (ctx *searchContext) sortRootMoves(moves []Move) {
	var child = ctx.Next()
	var list = make([]orderedMove, len(moves))
	for i, m := range moves {
		ctx.position.MakeMove(m, child.position)
		var score = -child.Quiescence(-VALUE_INFINITE, VALUE_INFINITE, 1)
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

func HashStorePV(ctx *searchContext, depth, score int, pv []Move) {
	for _, move := range pv {
		ctx.engine.transTable.Update(ctx.position, depth,
			ValueToTT(score, ctx.height), Lower|Upper, move)
		var child = ctx.Next()
		ctx.position.MakeMove(move, child.position)
		depth = ctx.NewDepth(depth, child)
		if depth <= 0 {
			break
		}
		ctx = child
	}
}

func (ctx *searchContext) AlphaBeta(alpha, beta, depth int) int {
	var newDepth, score int
	ctx.ClearPV()

	if ctx.height >= MAX_HEIGHT || ctx.IsDraw() {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return ctx.Quiescence(alpha, beta, 1)
	}

	var engine = ctx.engine
	engine.timeManager.IncNodes()

	if MateIn(ctx.height+1) <= alpha {
		return alpha
	}

	var position = ctx.position
	var hashMove = MoveEmpty

	if ttDepth, ttScore, ttType, ttMove, ok := engine.transTable.Read(position); ok {
		hashMove = ttMove
		if ttDepth >= depth {
			ttScore = ValueFromTT(ttScore, ctx.height)
			if ttScore >= beta && (ttType&Lower) != 0 {
				return beta
			}
			if ttScore <= alpha && (ttType&Upper) != 0 {
				return alpha
			}
		}
	}

	var isCheck = position.IsCheck()

	var child = ctx.Next()
	if depth >= 2 && !isCheck && position.LastMove != MoveEmpty &&
		beta < VALUE_MATE_IN_MAX_HEIGHT &&
		!IsLateEndgame(position, position.WhiteMove) {
		newDepth = depth - 4
		position.MakeNullMove(child.position)
		if newDepth <= 0 {
			score = -child.Quiescence(-beta, -(beta - 1), 1)
		} else {
			score = -child.AlphaBeta(-beta, -(beta - 1), newDepth)
		}
		if score >= beta && score < VALUE_MATE_IN_MAX_HEIGHT {
			return beta
		}
	}

	if depth >= 4 && hashMove == MoveEmpty && beta > alpha+PawnValue/2 {
		//good test: position fen 8/pp6/2p5/P1P5/1P3k2/3K4/8/8 w - - 5 47
		ctx.AlphaBeta(alpha, beta, depth-2)
		hashMove = ctx.BestMove()
		ctx.ClearPV() //!
	}

	var moveSort = moveSort{
		ctx:   ctx,
		trans: hashMove,
	}

	var moveCount = 0
	ctx.quietsSearched = ctx.quietsSearched[:0]
	var staticEval = VALUE_INFINITE

	for {
		var move = moveSort.Next()
		if move == MoveEmpty {
			break
		}

		if position.MakeMove(move, child.position) {
			moveCount++

			newDepth = ctx.NewDepth(depth, child)
			var reduction = 0

			if !IsCaptureOrPromotion(move) && moveCount > 1 &&
				!isCheck && !child.position.IsCheck() &&
				move != ctx.killer1 && move != ctx.killer2 &&
				!IsPawnPush7th(move, position.WhiteMove) &&
				alpha > VALUE_MATED_IN_MAX_HEIGHT {

				if depth <= 2 {
					if staticEval == VALUE_INFINITE {
						staticEval = engine.evaluate(position)
					}
					if staticEval+PawnValue <= alpha {
						continue
					}
				}

				if !IsPawnAdvance(move, position.WhiteMove) {
					if moveCount > 16 {
						reduction = 3
					} else if moveCount > 9 {
						reduction = 2
					} else {
						reduction = 1
					}
					reduction = Min(depth-2, reduction)
				}
			}

			if !IsCaptureOrPromotion(move) {
				ctx.quietsSearched = append(ctx.quietsSearched, move)
			}

			if reduction > 0 {
				score = -child.AlphaBeta(-(alpha + 1), -alpha, depth-1-reduction)
				if score <= alpha {
					continue
				}
			}

			score = -child.AlphaBeta(-beta, -alpha, newDepth)

			if score > alpha {
				alpha = score
				ctx.ComposePV(move, child)
				if alpha >= beta {
					break
				}
			}
		}
	}

	if moveCount == 0 {
		if isCheck {
			return MatedIn(ctx.height)
		}
		return VALUE_DRAW
	}

	var bestMove = ctx.BestMove()

	if bestMove != MoveEmpty && !IsCaptureOrPromotion(bestMove) {
		engine.historyTable.Update(ctx, bestMove, depth)
	}

	var ttType = 0
	if bestMove != MoveEmpty {
		ttType |= Lower
	}
	if alpha < beta {
		ttType |= Upper
	}
	engine.transTable.Update(position, depth, ValueToTT(alpha, ctx.height), ttType, bestMove)

	return alpha
}

func (ctx *searchContext) Quiescence(alpha, beta, depth int) int {
	var engine = ctx.engine
	engine.timeManager.IncNodes()
	ctx.ClearPV()
	if ctx.height >= MAX_HEIGHT {
		return VALUE_DRAW
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
	var child = ctx.Next()
	for {
		var move = moveSort.Next()
		if move == MoveEmpty {
			break
		}
		var danger = IsDangerCapture(position, move)
		if !isCheck && !danger && !SEE_GE(position, move) {
			continue
		}
		if position.MakeMove(move, child.position) {
			moveCount++
			if !isCheck && !danger && !child.position.IsCheck() &&
				eval+MoveValue(move)+PawnValue <= alpha {
				continue
			}
			var score = -child.Quiescence(-beta, -alpha, depth-1)
			if score > alpha {
				alpha = score
				ctx.ComposePV(move, child)
				if score >= beta {
					break
				}
			}
		}
	}
	if isCheck && moveCount == 0 {
		return MatedIn(ctx.height)
	}
	return alpha
}

func (ctx *searchContext) NewDepth(depth int, child *searchContext) int {
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

	if IsPawnPush7th(move, p.WhiteMove) && SEE_GE(p, move) {
		return depth
	}

	return depth - 1
}
