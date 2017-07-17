package engine

type SearchService struct {
	MoveOrderService      *MoveOrderService
	TTable                *TranspositionTable
	Evaluate              EvaluationFunc
	TimeControlStrategy   TimeControlStrategy
	DegreeOfParallelism   int
	UseExperimentSettings bool
	historyKeys           []uint64
	tm                    *TimeManagement
	stacks                []*SearchStack
}

func (this *SearchService) Search(searchParams SearchParams) (result SearchInfo) {
	var p = searchParams.Positions[len(searchParams.Positions)-1]
	this.tm = NewTimeManagement(searchParams.Limits, this.TimeControlStrategy,
		p.WhiteMove, searchParams.Context)

	this.historyKeys = PositionsToHistoryKeys(searchParams.Positions)
	this.MoveOrderService.Clear()
	if this.TTable != nil {
		this.TTable.PrepareNewSearch()
	}

	var ss = CreateStack()
	ss.Position = p
	ss.MoveList.GenerateMoves(p)
	ss.MoveList.FilterLegalMoves(p)

	if ss.MoveList.Count == 0 {
		return
	}

	result.MainLine = []Move{ss.MoveList.Items[0].Move}

	if ss.MoveList.Count == 1 {
		return
	}

	this.MoveOrderService.NoteMoves(ss, MoveEmpty)
	ss.MoveList.SortMoves()

	if this.DegreeOfParallelism <= 1 {
		result = this.IterateSearch(ss, searchParams.Progress)
	} else {
		result = this.IterateSearchParallel(ss, searchParams.Progress)
	}

	if len(result.MainLine) == 0 {
		result.MainLine = []Move{ss.MoveList.Items[0].Move}
	}
	result.Time = this.tm.ElapsedMilliseconds()
	result.Nodes = this.tm.Nodes()

	return
}

func (this *SearchService) IterateSearch(ss *SearchStack,
	progress func(SearchInfo)) (result SearchInfo) {
	defer func() {
		var r = recover()
		if r != nil && r != searchTimeout {
			panic(r)
		}
	}()

	var child = ss.Next

	const height = 0
	const beta = VALUE_INFINITE
	var p = ss.Position
	var ml = ss.MoveList
	for depth := 2; depth <= MAX_HEIGHT; depth++ {
		var prevScore = result.Score
		var alpha = -VALUE_INFINITE
		var bestMoveIndex = 0
		for i := 0; i < ml.Count; i++ {
			var move = ml.Items[i].Move
			p.MakeMove(move, child.Position)
			this.tm.IncNodes()
			var newDepth = NewDepth(depth, child)

			if alpha > VALUE_MATED_IN_MAX_HEIGHT &&
				-this.AlphaBeta(child, -(alpha+1), -alpha, newDepth, height+1, true) <= alpha {
				continue
			}
			var score = -this.AlphaBeta(child, -beta, -alpha, newDepth, height+1, false)

			if score > alpha {
				alpha = score
				result = SearchInfo{
					Depth:    depth,
					Score:    score,
					MainLine: append([]Move{move}, child.PrincipalVariation...),
					Time:     this.tm.ElapsedMilliseconds(),
					Nodes:    this.tm.Nodes(),
				}
				if progress != nil {
					progress(result)
				}
				bestMoveIndex = i
			}
		}
		if alpha >= MateIn(depth) || alpha <= MatedIn(depth) {
			break
		}
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && this.tm.IsSoftTimeout() {
			break
		}
		ml.MoveToBegin(bestMoveIndex)
	}
	return
}

func (this *SearchService) AlphaBeta(ss *SearchStack, alpha, beta, depth, height int,
	allowPrunings bool) int {
	var newDepth, score int
	ss.ClearPV()

	if height >= MAX_HEIGHT || IsDraw(ss, this.historyKeys) {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return this.Quiescence(ss, alpha, beta, 1, height)
	}

	this.tm.PanicOnHardTimeout()

	beta = min(beta, MateIn(height+1))
	if alpha >= beta {
		return alpha
	}

	var position = ss.Position
	var hashMove = MoveEmpty

	if this.TTable != nil {
		var ttDepth, ttScore, ttType, ttMove, ok = this.TTable.Read(position)
		if ok {
			hashMove = ttMove
			if ttDepth >= depth {
				ttScore = ValueFromTT(ttScore, height)
				if ttScore >= beta && (ttType&Lower) != 0 {
					return beta
				}
				if ttScore <= alpha && (ttType&Upper) != 0 {
					return alpha
				}
			}
		}
	}

	var isCheck = position.IsCheck()
	var lateEndgame = IsLateEndgame(position, position.WhiteMove)

	if depth <= 1 && !isCheck && allowPrunings {
		var eval = this.Evaluate(position)
		if eval+PawnValue <= alpha {
			return this.Quiescence(ss, alpha, beta, 1, height)
		}
		if eval-PawnValue >= beta && !lateEndgame &&
			!HasPawnOn7th(position, !position.WhiteMove) {
			return beta
		}
	}

	if depth >= 2 && !isCheck && allowPrunings &&
		beta < VALUE_MATE_IN_MAX_HEIGHT && !lateEndgame {
		newDepth = depth - 4
		position.MakeNullMove(ss.Next.Position)
		if newDepth <= 0 {
			score = -this.Quiescence(ss.Next, -beta, -(beta - 1), 1, height+1)
		} else {
			score = -this.AlphaBeta(ss.Next, -beta, -(beta - 1), newDepth, height+1, false)
		}
		if score >= beta {
			return beta
		}
	}

	if depth >= 4 && hashMove == MoveEmpty {
		newDepth = depth - 2
		this.AlphaBeta(ss, alpha, beta, newDepth, height, false)
		hashMove = ss.BestMove()
		ss.ClearPV() //!
	}

	ss.MoveList.GenerateMoves(position)
	this.MoveOrderService.NoteMoves(ss, hashMove)
	var moveCount = 0
	ss.QuietsSearched = ss.QuietsSearched[:0]

	for i := 0; i < ss.MoveList.Count; i++ {
		var move = ss.MoveList.ElementAt(i)

		if position.MakeMove(move, ss.Next.Position) {
			this.tm.IncNodes()
			moveCount++

			newDepth = NewDepth(depth, ss.Next)

			if !IsCaptureOrPromotion(move) {
				ss.QuietsSearched = append(ss.QuietsSearched, move)
			}

			if depth >= 4 && !isCheck && !ss.Next.Position.IsCheck() &&
				!lateEndgame && alpha > VALUE_MATED_IN_MAX_HEIGHT &&
				len(ss.QuietsSearched) > 4 && !IsActiveMove(position, move) {
				score = -this.AlphaBeta(ss.Next, -(alpha + 1), -alpha, depth-2, height+1, true)
				if score <= alpha {
					continue
				}
			}

			score = -this.AlphaBeta(ss.Next, -beta, -alpha, newDepth, height+1, true)

			if score > alpha {
				alpha = score
				ss.ComposePV(move)
				if alpha >= beta {
					break
				}
			}
		}
	}

	if moveCount == 0 {
		if isCheck {
			return MatedIn(height)
		}
		return VALUE_DRAW
	}

	var bestMove = ss.BestMove()

	if bestMove != MoveEmpty && !IsCaptureOrPromotion(bestMove) {
		this.MoveOrderService.UpdateHistory(ss, bestMove, depth)
	}

	if this.TTable != nil {
		var ttType = 0
		if bestMove != MoveEmpty {
			ttType |= Lower
		}
		if alpha < beta {
			ttType |= Upper
		}
		this.TTable.Update(position, depth, ValueToTT(alpha, height), ttType, bestMove)
	}

	return alpha
}

func (this *SearchService) Quiescence(ss *SearchStack, alpha, beta, depth, height int) int {
	this.tm.PanicOnHardTimeout()
	ss.ClearPV()
	if height >= MAX_HEIGHT {
		return VALUE_DRAW
	}
	var position = ss.Position
	var isCheck = position.IsCheck()
	var eval = 0
	if !isCheck {
		eval = this.Evaluate(position)
		if eval > alpha {
			alpha = eval
		}
		if eval >= beta {
			return alpha
		}
	}
	if isCheck {
		ss.MoveList.GenerateMoves(position)
	} else {
		ss.MoveList.GenerateCaptures(position, depth > 0)
	}
	this.MoveOrderService.NoteMoves(ss, MoveEmpty)
	var moveCount = 0
	for i := 0; i < ss.MoveList.Count; i++ {
		var move = ss.MoveList.ElementAt(i)
		if !isCheck {
			if eval+MoveValue(move)+PawnValue <= alpha &&
				!IsDirectCheck(position, move) {
				continue
			}
			if SEE(position, move) < 0 {
				continue
			}
		}

		if position.MakeMove(move, ss.Next.Position) {
			this.tm.IncNodes()
			moveCount++
			var score = -this.Quiescence(ss.Next, -beta, -alpha, depth-1, height+1)
			if score > alpha {
				alpha = score
				ss.ComposePV(move)
				if score >= beta {
					break
				}
			}
		}
	}
	if isCheck && moveCount == 0 {
		return MatedIn(height)
	}
	return alpha
}

func NewDepth(depth int, child *SearchStack) int {
	var p = child.Previous.Position
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
