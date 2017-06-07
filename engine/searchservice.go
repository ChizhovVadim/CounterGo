package engine

import (
	"sync"
	"sync/atomic"
	"time"
)

type SearchService struct {
	MoveOrderService      *MoveOrderService
	TTable                *TranspositionTable
	Evaluate              EvaluationFunc
	DegreeOfParallelism   int
	UseExperimentSettings bool
	nodes, maxNodes       int64
	historyKeys           []uint64
	ct                    *CancellationToken
}

func (this *SearchService) Search(searchParams SearchParams) (result SearchInfo) {
	var start = time.Now()
	if searchParams.CancellationToken != nil {
		this.ct = searchParams.CancellationToken
	} else {
		this.ct = &CancellationToken{}
	}

	var softLimit, hardLimit = ComputeThinkTime(searchParams.Limits,
		searchParams.Positions[len(searchParams.Positions)-1].WhiteMove)
	if hardLimit > 0 {
		var timer = time.AfterFunc(time.Duration(hardLimit)*time.Millisecond, func() {
			this.ct.Cancel()
		})
		defer timer.Stop()
	}

	this.nodes = 0
	this.maxNodes = int64(searchParams.Limits.Nodes)
	this.historyKeys = PositionsToHistoryKeys(searchParams.Positions)
	this.MoveOrderService.Clear()
	if this.TTable != nil {
		this.TTable.ClearStatistics()
		if searchParams.IsTraceEnabled {
			defer this.TTable.PrintStatistics()
		}
	}

	var p = searchParams.Positions[len(searchParams.Positions)-1]
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
				var local_alpha = alpha
				var local_i = i
				i++
				gate.Unlock()
				if local_i >= ss.MoveList.Count {
					return false
				}
				var move = ss.MoveList.Items[local_i].Move
				var ss2 = stacks[threadIndex]
				p.MakeMove(move, ss2.Next.Position)
				atomic.AddInt64(&this.nodes, 1)
				ss2.Next.Move = move
				var newDepth = NewDepth(depth, ss2)

				if local_alpha > VALUE_MATED_IN_MAX_HEIGHT &&
					-this.AlphaBeta(ss2.Next, -(local_alpha+1), -local_alpha, newDepth, true) <= local_alpha {
					return true
				}

				var score = -this.AlphaBeta(ss2.Next, -beta, -local_alpha, newDepth, false)
				gate.Lock()
				if score > alpha {
					alpha = score
					result.MainLine = append([]Move{move}, ss2.Next.PrincipalVariation...)
					result.Depth = depth
					result.Score = score
					result.Time = int64(time.Since(start) / time.Millisecond)
					result.Nodes = this.nodes
					if searchParams.Progress != nil {
						searchParams.Progress(result)
					}
					bestMoveIndex = local_i
				}
				gate.Unlock()
				return true
			})
		if alpha >= MateIn(depth) || alpha <= MatedIn(depth) {
			break
		}
		if softLimit > 0 &&
			AbsDelta(prevScore, alpha) <= PawnValue/2 &&
			time.Since(start) >= time.Duration(softLimit)*time.Millisecond {
			break
		}
		ss.MoveList.MoveToBegin(bestMoveIndex)
	}

	result.Time = int64(time.Since(start) / time.Millisecond)
	result.Nodes = this.nodes

	return
}

func (this *SearchService) AlphaBeta(ss *SearchStack, alpha, beta, depth int,
	allowPrunings bool) int {
	var newDepth, score int
	ss.ClearPV()

	if ss.Height >= MAX_HEIGHT || IsDraw(ss, this.historyKeys) {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return this.Quiescence(ss, alpha, beta, 1)
	}

	PanicOnTimeout(this)

	beta = min(beta, MateIn(ss.Height+1))
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
				ttScore = ValueFromTT(ttScore, ss.Height)
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
			return this.Quiescence(ss, alpha, beta, 1)
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
		ss.Next.Move = MoveEmpty
		if newDepth <= 0 {
			score = -this.Quiescence(ss.Next, -beta, -(beta - 1), 1)
		} else {
			score = -this.AlphaBeta(ss.Next, -beta, -(beta - 1), newDepth, false)
		}
		if score >= beta {
			return beta
		}
	}

	if depth >= 6 && hashMove == MoveEmpty {
		newDepth = depth - 2
		this.AlphaBeta(ss, alpha, beta, newDepth, false)
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
			atomic.AddInt64(&this.nodes, 1)
			moveCount++

			ss.Next.Move = move

			newDepth = NewDepth(depth, ss)

			if !IsCaptureOrPromotion(move) {
				ss.QuietsSearched = append(ss.QuietsSearched, move)
			}

			if depth >= 4 && !isCheck && !ss.Next.Position.IsCheck() &&
				!lateEndgame && alpha > VALUE_MATED_IN_MAX_HEIGHT &&
				len(ss.QuietsSearched) > 4 && !IsActiveMove(position, move) {
				score = -this.AlphaBeta(ss.Next, -(alpha + 1), -alpha, depth-2, true)
				if score <= alpha {
					continue
				}
			}

			score = -this.AlphaBeta(ss.Next, -beta, -alpha, newDepth, true)

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
			return MatedIn(ss.Height)
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
		this.TTable.Update(position, depth, ValueToTT(alpha, ss.Height), ttType, bestMove)
	}

	return alpha
}

func (this *SearchService) Quiescence(ss *SearchStack, alpha, beta, depth int) int {
	PanicOnTimeout(this)
	ss.ClearPV()
	if ss.Height >= MAX_HEIGHT {
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
			atomic.AddInt64(&this.nodes, 1)
			moveCount++
			var score = -this.Quiescence(ss.Next, -beta, -alpha, depth-1)
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
		return MatedIn(ss.Height)
	}
	return alpha
}

func PanicOnTimeout(ss *SearchService) {
	if ss.ct.IsCancellationRequested() ||
		(ss.maxNodes != 0 && ss.nodes >= ss.maxNodes) {
		panic(searchTimeout)
	}
}

func NewDepth(depth int, ss *SearchStack) int {
	if ss.Move != MoveEmpty &&
		ss.Move.To() == ss.Next.Move.To() &&
		ss.Next.Move.CapturedPiece() > Pawn &&
		ss.Move.CapturedPiece() > Pawn &&
		SEE(ss.Position, ss.Next.Move) >= 0 {
		return depth
	}

	if ss.Next.Position.IsCheck() &&
		(depth <= 1 || SEE(ss.Position, ss.Next.Move) >= 0) {
		return depth
	}

	if IsPawnPush7th(ss.Next.Move, ss.Position.WhiteMove) &&
		SEE(ss.Position, ss.Next.Move) >= 0 {
		return depth
	}

	return depth - 1
}
