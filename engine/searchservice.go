package engine

import "sync"

func RecoverFromSearchTimeout() {
	var r = recover()
	if r != nil && r != searchTimeout {
		panic(r)
	}
}

func (ctx *searchContext) IterateSearch(progress func(SearchInfo)) (result SearchInfo) {
	defer RecoverFromSearchTimeout()
	var engine = ctx.Engine
	defer func() {
		result.Time = engine.timeManager.ElapsedMilliseconds()
		result.Nodes = engine.timeManager.Nodes()
	}()

	var p = ctx.Position
	ctx.MoveList.GenerateMoves(p)
	ctx.MoveList.FilterLegalMoves(p)
	if ctx.MoveList.Count == 0 {
		return
	}
	result.MainLine = []Move{ctx.MoveList.Items[0].Move}
	if ctx.MoveList.Count == 1 {
		return
	}
	{
		var child = ctx.Next()
		for i := 0; i < ctx.MoveList.Count; i++ {
			var item = &ctx.MoveList.Items[i]
			p.MakeMove(item.Move, child.Position)
			item.Score = -child.Quiescence(-VALUE_INFINITE, VALUE_INFINITE, 1)
		}
		ctx.MoveList.SortMoves()
	}

	const height = 0
	const beta = VALUE_INFINITE
	var gate sync.Mutex
	for depth := 2; depth <= MAX_HEIGHT; depth++ {
		var prevScore = result.Score
		var alpha = -VALUE_INFINITE
		var bestMoveIndex = 0
		{
			var child = ctx.Next()
			var move = ctx.MoveList.Items[0].Move
			p.MakeMove(move, child.Position)
			var newDepth = ctx.NewDepth(depth, child)
			var score = -child.AlphaBeta(-beta, -alpha, newDepth, false)
			alpha = score
			result = SearchInfo{
				Depth:    depth,
				Score:    score,
				MainLine: append([]Move{move}, child.PrincipalVariation...),
				Time:     engine.timeManager.ElapsedMilliseconds(),
				Nodes:    engine.timeManager.Nodes(),
			}
			if progress != nil {
				progress(result)
			}
		}
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
				if localIndex >= ctx.MoveList.Count {
					return
				}
				var move = ctx.MoveList.Items[localIndex].Move
				p.MakeMove(move, child.Position)
				var newDepth = ctx.NewDepth(depth, child)
				var score = -child.AlphaBeta(-(localAlpha + 1), -localAlpha, newDepth, true)
				if score > localAlpha {
					score = -child.AlphaBeta(-beta, -localAlpha, newDepth, false)
				}
				gate.Lock()
				if score > alpha {
					alpha = score
					result = SearchInfo{
						Depth:    depth,
						Score:    score,
						MainLine: append([]Move{move}, child.PrincipalVariation...),
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
		if alpha >= MateIn(depth) || alpha <= MatedIn(depth) {
			break
		}
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && engine.timeManager.IsSoftTimeout() {
			break
		}
		ctx.MoveList.MoveToBegin(bestMoveIndex)
	}
	return
}

func (ctx *searchContext) AlphaBeta(alpha, beta, depth int, allowPrunings bool) int {
	var newDepth, score int
	ctx.ClearPV()

	if ctx.Height >= MAX_HEIGHT || ctx.IsDraw() {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return ctx.Quiescence(alpha, beta, 1)
	}

	var engine = ctx.Engine
	engine.timeManager.IncNodes()

	beta = min(beta, MateIn(ctx.Height+1))
	if alpha >= beta {
		return alpha
	}

	var position = ctx.Position
	var hashMove = MoveEmpty

	if ttDepth, ttScore, ttType, ttMove, ok := engine.transTable.Read(position); ok {
		hashMove = ttMove
		if ttDepth >= depth {
			ttScore = ValueFromTT(ttScore, ctx.Height)
			if ttScore >= beta && (ttType&Lower) != 0 {
				return beta
			}
			if ttScore <= alpha && (ttType&Upper) != 0 {
				return alpha
			}
		}
	}

	var isCheck = position.IsCheck()
	var lateEndgame = IsLateEndgame(position, position.WhiteMove)

	if depth <= 1 && !isCheck && allowPrunings {
		var eval = engine.evaluate(position)
		if eval+PawnValue <= alpha {
			return ctx.Quiescence(alpha, beta, 1)
		}
		if eval-PawnValue >= beta && !lateEndgame &&
			!HasPawnOn7th(position, !position.WhiteMove) {
			return beta
		}
	}

	var child = ctx.Next()
	if depth >= 2 && !isCheck && allowPrunings &&
		beta < VALUE_MATE_IN_MAX_HEIGHT && !lateEndgame {
		newDepth = depth - 4
		position.MakeNullMove(child.Position)
		if newDepth <= 0 {
			score = -child.Quiescence(-beta, -(beta - 1), 1)
		} else {
			score = -child.AlphaBeta(-beta, -(beta - 1), newDepth, false)
		}
		if score >= beta {
			return beta
		}
	}

	if depth >= 4 && hashMove == MoveEmpty {
		newDepth = depth - 2
		ctx.AlphaBeta(alpha, beta, newDepth, false)
		hashMove = ctx.BestMove()
		ctx.ClearPV() //!
	}

	ctx.InitMoves(hashMove)
	var moveCount = 0
	ctx.QuietsSearched = ctx.QuietsSearched[:0]

	for {
		var move = ctx.NextMove()
		if move == MoveEmpty {
			break
		}

		if position.MakeMove(move, child.Position) {
			moveCount++

			newDepth = ctx.NewDepth(depth, child)

			if !IsCaptureOrPromotion(move) {
				ctx.QuietsSearched = append(ctx.QuietsSearched, move)
			}

			if depth >= 4 && !isCheck && !child.Position.IsCheck() &&
				!lateEndgame && alpha > VALUE_MATED_IN_MAX_HEIGHT &&
				len(ctx.QuietsSearched) > 4 && !IsActiveMove(position, move) {
				score = -child.AlphaBeta(-(alpha + 1), -alpha, depth-2, true)
				if score <= alpha {
					continue
				}
			}

			score = -child.AlphaBeta(-beta, -alpha, newDepth, true)

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
			return MatedIn(ctx.Height)
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
	engine.transTable.Update(position, depth, ValueToTT(alpha, ctx.Height), ttType, bestMove)

	return alpha
}

func (ctx *searchContext) Quiescence(alpha, beta, depth int) int {
	var engine = ctx.Engine
	engine.timeManager.IncNodes()
	ctx.ClearPV()
	if ctx.Height >= MAX_HEIGHT {
		return VALUE_DRAW
	}
	var position = ctx.Position
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
	ctx.InitQMoves(depth > 0)
	var moveCount = 0
	var child = ctx.Next()
	for {
		var move = ctx.NextMove()
		if move == MoveEmpty {
			break
		}

		if !isCheck {
			if eval+MoveValue(move)+PawnValue <= alpha &&
				!IsDirectCheck(position, move) {
				continue
			}
			if SEE(position, move) < 0 {
				continue
			}
		}

		if position.MakeMove(move, child.Position) {
			moveCount++
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
		return MatedIn(ctx.Height)
	}
	return alpha
}

func (ctx *searchContext) NewDepth(depth int, child *searchContext) int {
	var p = ctx.Position
	var prevMove = p.LastMove
	var move = child.Position.LastMove
	var givesCheck = child.Position.IsCheck()

	if prevMove != MoveEmpty &&
		prevMove.To() == move.To() &&
		move.CapturedPiece() > Pawn &&
		prevMove.CapturedPiece() > Pawn &&
		SEE(p, move) >= 0 {
		return depth
	}

	if givesCheck && (depth <= 1 || SEE(p, move) >= 0) {
		return depth
	}

	if IsPawnPush7th(move, p.WhiteMove) && SEE(p, move) >= 0 {
		return depth
	}

	return depth - 1
}
