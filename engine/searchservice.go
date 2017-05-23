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
	isCancelRequest       bool
}

func (this *SearchService) Search(searchParams SearchParams) (result SearchInfo) {
	var start = time.Now()
	this.isCancelRequest = false

	var moveTime = ComputeThinkTime(searchParams.Limits,
		searchParams.Positions[len(searchParams.Positions)-1].WhiteMove)
	if moveTime > 0 {
		var timer = time.AfterFunc(time.Duration(moveTime)*time.Millisecond, func() {
			this.isCancelRequest = true
		})
		defer timer.Stop()
	}

	this.nodes = 0
	this.maxNodes = int64(searchParams.Limits.Nodes)
	this.MoveOrderService.Clear()
	if this.TTable != nil {
		this.TTable.ClearStatistics()
		if searchParams.IsTraceEnabled {
			defer this.TTable.PrintStatistics()
		}
	}

	var stacks = make([]*SearchStack, this.DegreeOfParallelism)
	for i := 0; i < len(stacks); i++ {
		stacks[i] = CreateStack(searchParams.Positions)
	}

	var ss = stacks[0]
	var p = ss.Position
	ss.MoveList.GenerateMoves(p)
	ss.MoveList.FilterLegalMoves(p)

	if ss.MoveList.Count == 0 {
		return
	}

	result.MainLine = &PrincipalVariation{ss.MoveList.Items[0].Move, nil}

	if ss.MoveList.Count == 1 {
		return
	}

	this.MoveOrderService.NoteMoves(ss, MoveEmpty)
	ss.MoveList.SortMoves()

	const beta = VALUE_INFINITE
	var gate sync.Mutex
	for depth := 2; depth <= MAX_HEIGHT; depth++ {
		var alpha = -VALUE_INFINITE
		var i = 0
		var bestMoveIndex = 0
		ParallelSearch(stacks,
			func(ss2 *SearchStack) bool {
				gate.Lock()
				var local_alpha = alpha
				var local_i = i
				i++
				gate.Unlock()
				if local_i >= ss.MoveList.Count {
					return false
				}
				var move = ss.MoveList.Items[local_i].Move
				p.MakeMove(move, ss2.Next.Position)
				atomic.AddInt64(&this.nodes, 1)
				ss2.Next.SkipNullMove = false
				ss2.Next.Move = move
				var newDepth = NewDepth(depth, ss2)

				if local_alpha > VALUE_MATED_IN_MAX_HEIGHT &&
					-this.AlphaBeta(ss2.Next, -(local_alpha+1), -local_alpha, newDepth) <= local_alpha {
					return true
				}

				var score = -this.AlphaBeta(ss2.Next, -beta, -local_alpha, newDepth)
				gate.Lock()
				if score > alpha {
					alpha = score
					result.MainLine = &PrincipalVariation{move, ss2.Next.PrincipalVariation}
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
		ss.MoveList.MoveToBegin(bestMoveIndex)
	}

	result.Time = int64(time.Since(start) / time.Millisecond)
	result.Nodes = this.nodes

	return
}

func (this *SearchService) AlphaBeta(ss *SearchStack, alpha, beta, depth int) int {
	var newDepth, score int
	ss.PrincipalVariation = nil

	if ss.Height >= MAX_HEIGHT || IsDraw(ss) {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return this.Quiescence(ss, alpha, beta, 1)
	}

	if this.isCancelRequest || (this.maxNodes != 0 && this.nodes >= this.maxNodes) {
		panic(searchTimeout)
	}

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

	if depth >= 2 && !isCheck && !ss.SkipNullMove &&
		beta < VALUE_MATE_IN_MAX_HEIGHT && !lateEndgame {
		newDepth = depth - 3
		position.MakeNullMove(ss.Next.Position)
		ss.Next.SkipNullMove = true
		ss.Next.Move = MoveEmpty
		if newDepth <= 0 {
			score = -this.Quiescence(ss.Next, -beta, -(beta - 1), 1)
		} else {
			score = -this.AlphaBeta(ss.Next, -beta, -(beta - 1), newDepth)
		}
		if score >= beta {
			return beta
		}
	}

	if depth >= 3 && hashMove == MoveEmpty {
		newDepth = depth - 2
		ss.SkipNullMove = true
		this.AlphaBeta(ss, alpha, beta, newDepth)
		if ss.PrincipalVariation != nil {
			hashMove = ss.PrincipalVariation.Move
			ss.PrincipalVariation = nil // !!
		}
	}

	ss.MoveList.GenerateMoves(position)
	this.MoveOrderService.NoteMoves(ss, hashMove)
	var moveCount = 0
	ss.QuietsSearched = ss.QuietsSearched[:0]
	var eval = VALUE_INFINITE

	for i := 0; i < ss.MoveList.Count; i++ {
		var move = ss.MoveList.ElementAt(i)

		if position.MakeMove(move, ss.Next.Position) {
			atomic.AddInt64(&this.nodes, 1)
			moveCount++

			ss.Next.SkipNullMove = false
			ss.Next.Move = move

			newDepth = NewDepth(depth, ss)

			if depth <= 2 &&
				!isCheck && !ss.Next.Position.IsCheck() &&
				!IsCaptureOrPromotion(move) &&
				!IsPawnPush(move, position.WhiteMove) &&
				move != hashMove {
				if eval == VALUE_INFINITE {
					eval = this.Evaluate(position)
				}
				var margin = let(depth <= 1, 100, 400)
				if eval+margin <= alpha {
					continue
				}
			}

			if !IsCaptureOrPromotion(move) {
				ss.QuietsSearched = append(ss.QuietsSearched, move)
			}

			/*if this.UseExperimentSettings &&
				depth >= 4 && !isCheck && !ss.Next.Position.IsCheck() &&
				alpha > VALUE_MATED_IN_MAX_HEIGHT && !lateEndgame &&
				!IsCaptureOrPromotion(move) &&
				!IsPawnPush(move, position.WhiteMove) &&
				!IsPassedPawnMove(position, move) &&
				len(ss.QuietsSearched) > 4 {
				var bound = alpha - 10
				score = -this.AlphaBeta(ss.Next, -(bound + 1), -bound, newDepth-1)
				if score <= bound {
					continue
				}
			}*/

			score = -this.AlphaBeta(ss.Next, -beta, -alpha, newDepth)

			if score > alpha {
				ss.PrincipalVariation =
					&PrincipalVariation{move, ss.Next.PrincipalVariation}
				alpha = score
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

	var bestMove = MoveEmpty
	if ss.PrincipalVariation != nil {
		bestMove = ss.PrincipalVariation.Move
	}

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
	if this.isCancelRequest || (this.maxNodes != 0 && this.nodes >= this.maxNodes) {
		panic(searchTimeout)
	}
	ss.PrincipalVariation = nil
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
		if !isCheck &&
			eval+MoveValue(move)+PawnValue <= alpha &&
			!IsDirectCheck(position, move) {
			continue
		}
		if !isCheck && SEE(position, move) < 0 {
			continue
		}
		if position.MakeMove(move, ss.Next.Position) {
			atomic.AddInt64(&this.nodes, 1)
			moveCount++
			var score = -this.Quiescence(ss.Next, -beta, -alpha, depth-1)
			if score > alpha {
				alpha = score
				ss.PrincipalVariation =
					&PrincipalVariation{move, ss.Next.PrincipalVariation}
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
