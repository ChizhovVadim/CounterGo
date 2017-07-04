package engine

import (
	"sync"
)

type SearchService struct {
	MoveOrderService      *MoveOrderService
	TTable                *TranspositionTable
	Evaluate              EvaluationFunc
	TimeControlStrategy   TimeControlStrategy
	DegreeOfParallelism   int
	UseExperimentSettings bool
	historyKeys           []uint64
	tm                    *TimeManagement
}

func (this *SearchService) Search(searchParams SearchParams) (result SearchInfo) {
	var p = searchParams.Positions[len(searchParams.Positions)-1]
	this.tm = NewTimeManagement(searchParams.Limits, this.TimeControlStrategy,
		p.WhiteMove, searchParams.CancellationToken)
	defer this.tm.Close()

	this.historyKeys = PositionsToHistoryKeys(searchParams.Positions)
	this.MoveOrderService.Clear()
	if this.TTable != nil {
		this.TTable.PrepareNewSearch()
	}

	var stacks = make([]*SearchStack, this.DegreeOfParallelism)
	for i := 0; i < len(stacks); i++ {
		stacks[i] = CreateStack(p)
	}

	var ss = stacks[0]
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

	const height = 0
	const beta = VALUE_INFINITE
	var gate sync.Mutex
	for depth := 2; depth <= MAX_HEIGHT; depth++ {
		var prevScore = result.Score
		var alpha = -VALUE_INFINITE
		var i = 0
		var bestMoveIndex = 0
		ParallelSearch(this.DegreeOfParallelism,
			func(threadIndex int) bool {
				gate.Lock()
				var localAlpha = alpha
				var localIndex = i
				i++
				gate.Unlock()
				if localIndex >= ss.MoveList.Count {
					return false
				}
				var move = ss.MoveList.Items[localIndex].Move
				var ss2 = stacks[threadIndex]
				p.MakeMove(move, ss2.Next.Position)
				this.tm.IncNodes()
				var newDepth = NewDepth(depth, ss2)

				if localAlpha > VALUE_MATED_IN_MAX_HEIGHT &&
					-this.AlphaBeta(ss2.Next, -(localAlpha+1), -localAlpha, newDepth, height+1, true) <= localAlpha {
					return true
				}

				var score = -this.AlphaBeta(ss2.Next, -beta, -localAlpha, newDepth, height+1, false)
				gate.Lock()
				if score > alpha {
					alpha = score
					result = SearchInfo{
						MainLine: append([]Move{move}, ss2.Next.PrincipalVariation...),
						Depth:    depth,
						Score:    score,
						Time:     this.tm.ElapsedMilliseconds(),
						Nodes:    this.tm.Nodes(),
					}
					if searchParams.Progress != nil {
						searchParams.Progress(result)
					}
					bestMoveIndex = localIndex
				}
				gate.Unlock()
				return true
			})
		if alpha >= MateIn(depth) || alpha <= MatedIn(depth) {
			break
		}
		if AbsDelta(prevScore, alpha) <= PawnValue/2 && this.tm.IsSoftTimeout() {
			break
		}
		ss.MoveList.MoveToBegin(bestMoveIndex)
	}

	result.Time = this.tm.ElapsedMilliseconds()
	result.Nodes = this.tm.Nodes()

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

			newDepth = NewDepth(depth, ss)

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

func NewDepth(depth int, ss *SearchStack) int {
	var prevMove = ss.Position.LastMove
	var move = ss.Next.Position.LastMove

	if prevMove != MoveEmpty &&
		prevMove.To() == move.To() &&
		move.CapturedPiece() > Pawn &&
		prevMove.CapturedPiece() > Pawn &&
		SEE(ss.Position, move) >= 0 {
		return depth
	}

	if ss.Next.Position.IsCheck() &&
		(depth <= 1 || SEE(ss.Position, move) >= 0) {
		return depth
	}

	if IsPawnPush7th(move, ss.Position.WhiteMove) &&
		SEE(ss.Position, move) >= 0 {
		return depth
	}

	return depth - 1
}
