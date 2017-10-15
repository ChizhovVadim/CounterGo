package engine

import "sync"

type SearchService struct {
	HistoryTable          HistoryTable
	TTable                *TranspositionTable
	Evaluate              EvaluationFunc
	TimeControlStrategy   TimeControlStrategy
	DegreeOfParallelism   int
	UseExperimentSettings bool
	historyKeys           []uint64
	tm                    *TimeManagement
	tree                  [][]SearchStack
}

func (this *SearchService) Search(searchParams SearchParams) (result SearchInfo) {
	var p = searchParams.Positions[len(searchParams.Positions)-1]
	this.tm = NewTimeManagement(searchParams.Limits, this.TimeControlStrategy,
		p.WhiteMove, searchParams.CancellationToken)
	defer this.tm.Close()

	this.historyKeys = PositionsToHistoryKeys(searchParams.Positions)
	this.HistoryTable.Clear()
	if this.TTable != nil {
		this.TTable.PrepareNewSearch()
	}

	if len(this.tree) != this.DegreeOfParallelism {
		this.tree = NewTree(this, this.DegreeOfParallelism)
	}

	for i := 0; i < len(this.tree); i++ {
		this.tree[i][0].Position = p
	}

	var ss = &this.tree[0][0]
	ss.MoveList.GenerateMoves(p)
	ss.MoveList.FilterLegalMoves(p)

	if ss.MoveList.Count == 0 {
		return
	}

	result.MainLine = []Move{ss.MoveList.Items[0].Move}

	if ss.MoveList.Count == 1 {
		return
	}

	var child = ss.Next()
	for i := 0; i < ss.MoveList.Count; i++ {
		var item = &ss.MoveList.Items[i]
		p.MakeMove(item.Move, child.Position)
		item.Score = -child.Quiescence(-VALUE_INFINITE, VALUE_INFINITE, 1)
	}
	ss.MoveList.SortMoves()

	result = this.IterateSearch(ss, searchParams.Progress)

	if len(result.MainLine) == 0 {
		result.MainLine = []Move{ss.MoveList.Items[0].Move}
	}
	result.Time = this.tm.ElapsedMilliseconds()
	result.Nodes = this.tm.Nodes()

	return
}

func RecoverFromSearchTimeout() {
	var r = recover()
	if r != nil && r != searchTimeout {
		panic(r)
	}
}

func (this *SearchService) IterateSearch(ss *SearchStack,
	progress func(SearchInfo)) (result SearchInfo) {
	defer RecoverFromSearchTimeout()

	const height = 0
	const beta = VALUE_INFINITE
	var gate sync.Mutex
	var p = ss.Position
	var ml = ss.MoveList
	for depth := 2; depth <= MAX_HEIGHT; depth++ {
		var prevScore = result.Score
		var alpha = -VALUE_INFINITE
		var bestMoveIndex = 0
		{
			var child = &this.tree[0][height+1]
			var move = ml.Items[0].Move
			p.MakeMove(move, child.Position)
			this.tm.IncNodes()
			var newDepth = ss.NewDepth(depth, child)
			var score = -child.AlphaBeta(-beta, -alpha, newDepth, false)
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
		}
		var index = 1
		ParallelDo(this.DegreeOfParallelism, func(threadIndex int) {
			defer RecoverFromSearchTimeout()
			var child = &this.tree[threadIndex][height+1]
			for {
				gate.Lock()
				var localAlpha = alpha
				var localIndex = index
				index++
				gate.Unlock()
				if localIndex >= ml.Count {
					return
				}
				var move = ml.Items[localIndex].Move
				p.MakeMove(move, child.Position)
				this.tm.IncNodes()
				var newDepth = ss.NewDepth(depth, child)
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
						Time:     this.tm.ElapsedMilliseconds(),
						Nodes:    this.tm.Nodes(),
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
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && this.tm.IsSoftTimeout() {
			break
		}
		ml.MoveToBegin(bestMoveIndex)
	}
	return
}

func (ss *SearchStack) AlphaBeta(alpha, beta, depth int, allowPrunings bool) int {
	var newDepth, score int
	ss.ClearPV()

	if ss.height >= MAX_HEIGHT || ss.IsDraw() {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return ss.Quiescence(alpha, beta, 1)
	}

	var searchService = ss.searchService
	searchService.tm.PanicOnHardTimeout()

	beta = min(beta, MateIn(ss.height+1))
	if alpha >= beta {
		return alpha
	}

	var position = ss.Position
	var hashMove = MoveEmpty

	if ttDepth, ttScore, ttType, ttMove, ok := searchService.TTable.Read(position); ok {
		hashMove = ttMove
		if ttDepth >= depth {
			ttScore = ValueFromTT(ttScore, ss.height)
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
		var eval = searchService.Evaluate(position)
		if eval+PawnValue <= alpha {
			return ss.Quiescence(alpha, beta, 1)
		}
		if eval-PawnValue >= beta && !lateEndgame &&
			!HasPawnOn7th(position, !position.WhiteMove) {
			return beta
		}
	}

	var child = ss.Next()
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
		ss.AlphaBeta(alpha, beta, newDepth, false)
		hashMove = ss.BestMove()
		ss.ClearPV() //!
	}

	ss.InitMoves(hashMove)
	var moveCount = 0
	ss.QuietsSearched = ss.QuietsSearched[:0]

	for {
		var move = ss.NextMove()
		if move == MoveEmpty {
			break
		}

		if position.MakeMove(move, child.Position) {
			searchService.tm.IncNodes()
			moveCount++

			newDepth = ss.NewDepth(depth, child)

			if !IsCaptureOrPromotion(move) {
				ss.QuietsSearched = append(ss.QuietsSearched, move)
			}

			if depth >= 4 && !isCheck && !child.Position.IsCheck() &&
				!lateEndgame && alpha > VALUE_MATED_IN_MAX_HEIGHT &&
				len(ss.QuietsSearched) > 4 && !IsActiveMove(position, move) {
				score = -child.AlphaBeta(-(alpha + 1), -alpha, depth-2, true)
				if score <= alpha {
					continue
				}
			}

			score = -child.AlphaBeta(-beta, -alpha, newDepth, true)

			if score > alpha {
				alpha = score
				ss.ComposePV(move, child)
				if alpha >= beta {
					break
				}
			}
		}
	}

	if moveCount == 0 {
		if isCheck {
			return MatedIn(ss.height)
		}
		return VALUE_DRAW
	}

	var bestMove = ss.BestMove()

	if bestMove != MoveEmpty && !IsCaptureOrPromotion(bestMove) {
		searchService.HistoryTable.Update(ss, bestMove, depth)
	}

	var ttType = 0
	if bestMove != MoveEmpty {
		ttType |= Lower
	}
	if alpha < beta {
		ttType |= Upper
	}
	searchService.TTable.Update(position, depth, ValueToTT(alpha, ss.height), ttType, bestMove)

	return alpha
}

func (ss *SearchStack) Quiescence(alpha, beta, depth int) int {
	var searchService = ss.searchService
	searchService.tm.PanicOnHardTimeout()
	ss.ClearPV()
	if ss.height >= MAX_HEIGHT {
		return VALUE_DRAW
	}
	var position = ss.Position
	var isCheck = position.IsCheck()
	var eval = 0
	if !isCheck {
		eval = searchService.Evaluate(position)
		if eval > alpha {
			alpha = eval
		}
		if eval >= beta {
			return alpha
		}
	}
	ss.InitQMoves(depth > 0)
	var moveCount = 0
	var child = ss.Next()
	for {
		var move = ss.NextMove()
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
			searchService.tm.IncNodes()
			moveCount++
			var score = -child.Quiescence(-beta, -alpha, depth-1)
			if score > alpha {
				alpha = score
				ss.ComposePV(move, child)
				if score >= beta {
					break
				}
			}
		}
	}
	if isCheck && moveCount == 0 {
		return MatedIn(ss.height)
	}
	return alpha
}

func (ss *SearchStack) NewDepth(depth int, child *SearchStack) int {
	var p = ss.Position
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
