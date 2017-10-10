package engine

import (
	"sync"
)

func ParallelDo(degreeOfParallelism int, body func(threadIndex int)) {
	var wg sync.WaitGroup
	for i := 1; i < degreeOfParallelism; i++ {
		wg.Add(1)
		go func(threadIndex int) {
			body(threadIndex)
			wg.Done()
		}(i)
	}
	body(0)
	wg.Wait()
}

func (this *SearchService) IterateSearchParallel(ss *SearchStack,
	progress func(SearchInfo)) (result SearchInfo) {
	defer func() {
		var r = recover()
		if r != nil && r != searchTimeout {
			panic(r)
		}
	}()

	var p = ss.Position
	var ml = ss.MoveList

	this.stacks = make([]*SearchStack, this.DegreeOfParallelism)
	for i := 1; i < len(this.stacks); i++ {
		this.stacks[i] = CreateStack()
	}

	const height = 0
	const beta = VALUE_INFINITE
	var gate sync.Mutex
	for depth := 2; depth <= MAX_HEIGHT; depth++ {
		var prevScore = result.Score
		var alpha = -VALUE_INFINITE
		var bestMoveIndex = 0
		{
			var child = ss.Next
			var move = ml.Items[0].Move
			p.MakeMove(move, child.Position)
			this.tm.IncNodes()
			var newDepth = NewDepth(depth, child)
			var score = -this.AlphaBetaParallel(child, -beta, -alpha, newDepth, height+1, false)
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
			defer func() {
				var r = recover()
				if r != nil && r != searchTimeout {
					panic(r)
				}
			}()

			var child *SearchStack
			if threadIndex == 0 {
				child = ss.Next
			} else {
				child = this.stacks[threadIndex]
				child.Previous = ss
			}
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
				var newDepth = NewDepth(depth, child)

				if localAlpha > VALUE_MATED_IN_MAX_HEIGHT &&
					-this.AlphaBeta(child, -(localAlpha+1), -localAlpha, newDepth, height+1, true) <= localAlpha {
					continue
				}
				var score = -this.AlphaBeta(child, -beta, -localAlpha, newDepth, height+1, false)

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
		ss.MoveList.MoveToBegin(bestMoveIndex)
	}
	return
}

func (this *SearchService) AlphaBetaParallel(ss *SearchStack, alpha, beta, depth, height int,
	allowPrunings bool) int {
	ss.ClearPV()

	if height >= MAX_HEIGHT || IsDraw(ss, this.historyKeys) {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return this.Quiescence(ss, alpha, beta, 1, height)
	}

	if depth <= 4 {
		return this.AlphaBeta(ss, alpha, beta, depth, height, allowPrunings)
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
		position.MakeNullMove(ss.Next.Position)
		var newDepth = depth - 4
		var score int
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
		var newDepth = depth - 2
		this.AlphaBetaParallel(ss, alpha, beta, newDepth, height, false)
		hashMove = ss.BestMove()
		ss.ClearPV() //!
	}

	ss.MoveList.GenerateMoves(position)
	this.MoveOrderService.NoteMoves(ss, hashMove)
	ss.QuietsSearched = ss.QuietsSearched[:0]

	var i int
	{
		var child = ss.Next
		var moveCount = 0

		for i = 0; i < ss.MoveList.Count; i++ {
			var move = ss.MoveList.ElementAt(i)

			if !position.MakeMove(move, child.Position) {
				continue
			}

			this.tm.IncNodes()
			moveCount++

			var newDepth = NewDepth(depth, child)

			if !IsCaptureOrPromotion(move) {
				ss.QuietsSearched = append(ss.QuietsSearched, move)
			}

			var score = -this.AlphaBetaParallel(child, -beta, -alpha, newDepth, height+1, true)

			if score > alpha {
				alpha = score
				ss.ComposePV(move)
			}

			i++
			break //search first legal move before parallel search
		}

		if moveCount == 0 {
			if isCheck {
				return MatedIn(height)
			}
			return VALUE_DRAW
		}
	}

	var gate sync.Mutex
	ParallelDo(this.DegreeOfParallelism, func(threadIndex int) {
		defer func() {
			var r = recover()
			if r != nil && r != searchTimeout {
				panic(r)
			}
		}()

		var child *SearchStack
		if threadIndex == 0 {
			child = ss.Next
		} else {
			child = this.stacks[threadIndex]
			child.Previous = ss
		}

		for {
			gate.Lock()
			var localAlpha = alpha
			var move = MoveEmpty
			if i < ss.MoveList.Count {
				move = ss.MoveList.ElementAt(i)
				i++
			}
			var quiets = len(ss.QuietsSearched)
			gate.Unlock()

			if move == MoveEmpty {
				return
			}
			if localAlpha >= beta {
				return
			}
			if !position.MakeMove(move, child.Position) {
				continue
			}

			this.tm.IncNodes()
			var newDepth = NewDepth(depth, child)

			var score int
			if depth >= 4 && !isCheck && !child.Position.IsCheck() &&
				!lateEndgame && localAlpha > VALUE_MATED_IN_MAX_HEIGHT &&
				quiets > 4 && !IsActiveMove(position, move) {
				score = -this.AlphaBeta(child, -(localAlpha + 1), -localAlpha, depth-2, height+1, true)
			} else {
				score = localAlpha + 1
			}

			if score > localAlpha {
				score = -this.AlphaBeta(child, -beta, -localAlpha, newDepth, height+1, true)
			}

			gate.Lock()
			if score > alpha {
				alpha = score
				ss.PrincipalVariation = append(append(ss.PrincipalVariation[:0], move), child.PrincipalVariation...)
			}
			if !IsCaptureOrPromotion(move) {
				ss.QuietsSearched = append(ss.QuietsSearched, move)
			}
			gate.Unlock()
		}
	})
	this.tm.PanicOnHardTimeout() //!

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
