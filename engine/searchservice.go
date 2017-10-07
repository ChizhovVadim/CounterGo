package engine

import (
	"math"
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
	stacks                []*SearchStack
}

func (this *SearchService) Search(searchParams SearchParams) (result SearchInfo) {
	var p = searchParams.Positions[len(searchParams.Positions)-1]
	this.tm = NewTimeManagement(searchParams.Limits, this.TimeControlStrategy,
		p.WhiteMove, searchParams.Context)

	this.historyKeys = PositionsToHistoryKeys(searchParams.Positions)
	this.MoveOrderService.Clear()

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

	result = this.IterateSearch(ss, searchParams.Progress)

	if len(result.MainLine) == 0 {
		result.MainLine = []Move{ss.MoveList.Items[0].Move}
	}
	result.Time = this.tm.ElapsedMilliseconds()
	result.Nodes = this.tm.Nodes()

	return
}

func (this *SearchService) IterateSearch(ss *SearchStack,
	progress func(SearchInfo)) (result SearchInfo) {
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
			var score = -this.AlphaBeta(child, -beta, -alpha, newDepth, height+1, false, 0)
			if IsCancelValue(score) {
				return
			}
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

				if localAlpha > VALUE_MATED_IN_MAX_HEIGHT {
					var score = -this.AlphaBeta(child, -(localAlpha + 1), -localAlpha, newDepth, height+1, true, 0)
					if IsCancelValue(score) {
						return
					}
					if score <= localAlpha {
						continue
					}
				}
				var score = -this.AlphaBeta(child, -beta, -localAlpha, newDepth, height+1, false, 0)
				if IsCancelValue(score) {
					return
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
		ss.MoveList.MoveToBegin(bestMoveIndex)
	}
	return
}

func (this *SearchService) AlphaBeta(ss *SearchStack, alpha, beta, depth, height int,
	allowPrunings bool, lmrReduction int) int {
	var newDepth, score int
	ss.ClearPV()

	if height >= MAX_HEIGHT || IsDraw(ss, this.historyKeys) {
		return VALUE_DRAW
	}

	if depth <= 0 {
		return this.Quiescence(ss, alpha, beta, 1, height)
	}

	if this.tm.IsHardTimeout() {
		return VALUE_CANCEL
	}

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
		this.tm.IncNodes()
		if newDepth <= 0 {
			score = -this.Quiescence(ss.Next, -beta, -(beta - 1), 1, height+1)
		} else {
			score = -this.AlphaBeta(ss.Next, -beta, -(beta - 1), newDepth, height+1, false, 0)
		}
		if IsCancelValue(score) {
			return score
		}
		if score >= beta {
			return beta
		}
	}

	if lmrReduction > 0 {
		var rbeta = min(VALUE_INFINITE, beta+10)
		score = this.AlphaBeta(ss, rbeta-1, rbeta, depth-lmrReduction, height, false, 0)
		if IsCancelValue(score) {
			return score
		}
		if score >= rbeta {
			return beta
		}
		ss.ClearPV()
	}

	if depth >= 4 && hashMove == MoveEmpty {
		newDepth = depth - 2
		score = this.AlphaBeta(ss, alpha, beta, newDepth, height, false, 0)
		if IsCancelValue(score) {
			return score
		}
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

			var lmrReduction = 0
			if depth >= 3 && !isCheck && !ss.Next.Position.IsCheck() &&
				!lateEndgame && alpha > VALUE_MATED_IN_MAX_HEIGHT &&
				moveCount > 1 && move != ss.KillerMove &&
				!IsCaptureOrPromotion(move) &&
				!IsPawnAdvance(move, position.WhiteMove) {
				//lmrReduction = 1
				lmrReduction = lateMoveReductions[min(31, depth)][min(63, moveCount)]
			}

			score = -this.AlphaBeta(ss.Next, -beta, -alpha, newDepth, height+1, true, lmrReduction)
			if IsCancelValue(score) {
				return score
			}

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
	if this.tm.IsHardTimeout() {
		return VALUE_CANCEL
	}
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
			if IsCancelValue(score) {
				return score
			}
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

var lateMoveReductions [32][64]int

func init() {
	// Late move reductions from Crafty
	const (
		LMR_rdepth = 1   /* leave 1 full ply after reductions    */
		LMR_min    = 1   /* minimum reduction 1 ply              */
		LMR_max    = 15  /* maximum reduction 15 plies           */
		LMR_db     = 1.8 /* depth is 1.8x as important as        */
		LMR_mb     = 1.0 /* moves searched in the formula.       */
		LMR_s      = 2.0 /* smaller numbers increase reductions. */
	)

	for d := 3; d < 32; d++ {
		for m := 1; m < 64; m++ {
			var reduction = int(math.Log(float64(d)*LMR_db) * math.Log(float64(m)*LMR_mb) / LMR_s)
			reduction = max(min(reduction, LMR_max), LMR_min)
			reduction = min(reduction, max(d-1-LMR_rdepth, 0))
			lateMoveReductions[d][m] = reduction
		}
	}
}
