package engine

import (
	"errors"

	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const pawnValue = 100

var errSearchTimeout = errors.New("search timeout")

type thread struct {
	idx           int
	evaluator     IUpdatableEvaluator
	sharedContext *SharedContext
	nodes         int
	rootDepth     int
	stack         [stackSize]struct {
		position       Position
		moveList       [MaxMoves]OrderedMove
		quietsSearched [MaxMoves]Move
		pv             pv
		staticEval     int
		killer1        Move
		killer2        Move
	}
	mainHistory         [8192]int16
	continuationHistory [1024][1024]int16
}

type pv struct {
	items [stackSize]Move
	size  int
}

func (t *thread) Clear() {
	t.clearHistory()
}

func (t *thread) iterativeDeepening(
	sharedContext *SharedContext,
) (result mainLine) {
	t.sharedContext = sharedContext
	defer func() {
		t.sharedContext = nil
	}()

	t.nodes = 0
	t.stack[0].position = sharedContext.game.Position
	t.evaluator.Init(&t.stack[0].position)
	for h := 0; h <= 2; h++ {
		t.stack[h].killer1 = MoveEmpty
		t.stack[h].killer2 = MoveEmpty
	}

	var ml = sharedContext.game.Position.GenerateLegalMoves()
	if len(ml) == 0 {
		return mainLine{}
	}

	// select random legal move
	result = mainLine{
		moves: []Move{ml[0]},
	}

	defer func() {
		if r := recover(); r != nil {
			if r == errSearchTimeout {
				return
			}
			panic(r)
		}
	}()

	for depth := 1; depth <= maxHeight; depth += 1 {
		var score = t.aspirationWindow(depth, result.score)
		result = mainLine{
			depth: depth,
			score: score,
			moves: t.stack[0].pv.clone(),
		}
		if len(result.moves) == 0 {
			panic("empty best line")
		}
		if t.idx == 0 {
			if depth >= 4 && len(ml) == 1 {
				break
			}
			sharedContext.OnIterationComplete(result)
		}
	}
	return result
}

func (t *thread) aspirationWindow(depth, prevScore int) int {
	if depth >= 5 && !(prevScore <= valueLoss || prevScore >= valueWin) {
		const Window = 25
		var alpha = Max(-valueInfinity, prevScore-Window)
		var beta = Min(valueInfinity, prevScore+Window)
		var score = t.searchRoot(alpha, beta, depth)
		if score > alpha && score < beta {
			return score
		}
		if score >= beta {
			beta = valueInfinity
		}
		if score <= alpha {
			alpha = -valueInfinity
		}
		score = t.searchRoot(alpha, beta, depth)
		if score > alpha && score < beta {
			return score
		}
	}
	return t.searchRoot(-valueInfinity, valueInfinity, depth)
}

func (t *thread) searchRoot(alpha, beta, depth int) int {
	t.rootDepth = depth
	return t.alphaBeta(alpha, beta, depth, 0, 0)
}

// main search method
func (t *thread) alphaBeta(alpha, beta, depth, height int, skipMove Move) int {
	if depth <= 0 {
		return t.quiescence(alpha, beta, height)
	}
	t.clearPV(height)

	var rootNode = height == 0
	var pvNode = beta != alpha+1
	var position = &t.stack[height].position
	var isCheck = position.IsCheck()
	var ttMoveIsSingular = false

	if !rootNode {
		if height >= maxHeight {
			return t.evaluator.EvaluateQuick(position)
		}
		if t.isRepeat(height) {
			return valueDraw
		}
		if isDraw(position) {
			return valueDraw
		}
		// mate distance pruning
		if winIn(height+1) <= alpha {
			return alpha
		}
		if lossIn(height+2) >= beta && !isCheck {
			return beta
		}
	}

	// transposition table
	var (
		ttDepth, ttValue, ttBound int
		ttMove                    Move
		ttHit                     bool
	)
	if skipMove == 0 {
		ttDepth, ttValue, ttBound, ttMove, ttHit = t.sharedContext.transTable.Read(position.Key)
	}
	if ttHit {
		ttValue = valueFromTT(ttValue, height)
		if ttDepth >= depth && !pvNode && !(position.LastMove == MoveEmpty) {
			if ttValue >= beta && (ttBound&boundLower) != 0 {
				if ttMove != MoveEmpty && !isCaptureOrPromotion(ttMove) {
					t.updateKiller(ttMove, height)
				}
				return ttValue
			}
			if ttValue <= alpha && (ttBound&boundUpper) != 0 {
				return ttValue
			}
		}
	}

	var staticEval = t.evaluator.EvaluateQuick(position)
	t.stack[height].staticEval = staticEval
	var improving = height < 2 || staticEval > t.stack[height-2].staticEval

	if height+2 <= maxHeight {
		t.stack[height+2].killer1 = MoveEmpty
		t.stack[height+2].killer2 = MoveEmpty
	}
	var child = &t.stack[height+1].position

	if !rootNode && skipMove == 0 {

		// reverse futility pruning
		if !pvNode && depth <= 8 && !isCheck {
			var score = staticEval - pawnValue*depth
			if score >= beta {
				return staticEval
			}
		}

		// null-move pruning
		if !pvNode && depth >= 2 && !isCheck &&
			position.LastMove != MoveEmpty &&
			(height <= 1 || t.stack[height-1].position.LastMove != MoveEmpty) &&
			beta < valueWin &&
			!(ttHit && ttValue < beta && (ttBound&boundUpper) != 0) &&
			!isLateEndgame(position, position.WhiteMove) &&
			staticEval >= beta {
			var reduction = 4 + depth/6 + Min(2, (staticEval-beta)/200)
			t.makeMove(MoveEmpty, height)
			var score = -t.alphaBeta(-beta, -(beta - 1), depth-reduction, height+1, 0)
			t.unmakeMove()
			if score >= beta {
				if score >= valueWin {
					score = beta
				}
				return score
			}
		}

		var probcutBeta = Min(valueWin-1, beta+150)
		if !pvNode && depth >= 5 && !isCheck &&
			beta > valueLoss && beta < valueWin &&
			!(ttHit && ttDepth >= depth-4 && ttValue < probcutBeta && (ttBound&boundUpper) != 0) {

			var mi = moveIteratorQS{
				position: position,
				buffer:   t.stack[height].moveList[:],
			}
			mi.Init()

			for mi.Reset(); ; {
				var move = mi.Next()
				if move == MoveEmpty {
					break
				}
				if !seeGEZero(position, move) {
					continue
				}
				if !t.makeMove(move, height) {
					continue
				}
				var score = -t.quiescence(-probcutBeta, -probcutBeta+1, height+1)
				if score >= probcutBeta {
					score = -t.alphaBeta(-probcutBeta, -probcutBeta+1, depth-4, height+1, 0)
				}
				t.unmakeMove()
				if score >= probcutBeta {
					return score
				}
			}
		}

		// singular extension
		if depth >= 8 &&
			ttHit && ttMove != MoveEmpty &&
			(ttBound&boundLower) != 0 && ttDepth >= depth-3 &&
			ttValue > valueLoss && ttValue < valueWin {
			var singularBeta = Max(-valueInfinity, ttValue-depth)
			var score = t.alphaBeta(singularBeta-1, singularBeta, depth/2, height, ttMove)
			ttMoveIsSingular = score < singularBeta
		}
	}

	var historyContext = t.getHistoryContext(height)

	var mi = t.initMoveIterator(height, ttMove)
	var killer1 = t.stack[height].killer1
	var killer2 = t.stack[height].killer2

	var movesSearched = 0
	var hasLegalMove = false
	var quietsSeen = 0

	var quietsSearched = t.stack[height].quietsSearched[:0]
	var bestMove Move

	var lmp = 5 + (depth-1)*depth
	if !improving {
		lmp /= 2
	}

	var best = -valueInfinity
	var oldAlpha = alpha

	for mi.Reset(); ; {
		var move = mi.Next()
		if move == MoveEmpty {
			break
		}
		if move == skipMove {
			continue
		}
		var isNoisy = isCaptureOrPromotion(move)
		if !isNoisy {
			quietsSeen++
		}

		if depth <= 8 && best > valueLoss && hasLegalMove && !isCheck && !rootNode {
			// late-move pruning
			if !(isNoisy ||
				move == killer1 ||
				move == killer2) &&
				quietsSeen > lmp {
				continue
			}

			// futility pruning
			if !(isNoisy ||
				move == killer1 ||
				move == killer2) &&
				staticEval+100+pawnValue*depth <= alpha {
				continue
			}

			// SEE pruning
			{
				var seeMargin int
				if isNoisy {
					seeMargin = Max(depth, (staticEval+pawnValue-alpha)/pawnValue)
				} else {
					seeMargin = depth / 2
				}
				if !SeeGE(position, move, -seeMargin) {
					continue
				}
			}
		}

		if !t.makeMove(move, height) {
			continue
		}
		hasLegalMove = true

		movesSearched++

		var extension, reduction int

		if child.IsCheck() && depth >= 3 {
			extension = 1
		}
		if move == ttMove && ttMoveIsSingular {
			extension = 1
		}

		if depth >= 3 && movesSearched > 1 &&
			!(isNoisy) {
			reduction = lmr(depth, movesSearched)
			if move == killer1 || move == killer2 {
				reduction--
			}
			if !isCheck {
				var history = historyContext.ReadTotal(move)
				reduction -= Max(-2, Min(2, history/5000))

				if !improving {
					reduction++
				}
			}
			if pvNode {
				reduction -= 2
			}
			if isCheck || child.IsCheck() {
				reduction--
			}
			reduction = Max(reduction, 0) + extension
			reduction = Max(0, Min(depth-2, reduction))
		}

		if !isNoisy {
			quietsSearched = append(quietsSearched, move)
		}

		var newDepth = depth - 1 + extension

		var score = alpha + 1
		// LMR
		if reduction > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth-reduction, height+1, 0)
		}
		// PVS
		if score > alpha && beta != alpha+1 && movesSearched > 1 && newDepth > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth, height+1, 0)
		}
		// full search
		if score > alpha {
			score = -t.alphaBeta(-beta, -alpha, newDepth, height+1, 0)
		}

		t.unmakeMove()

		if score > best {
			best = score
			bestMove = move
		}
		if score > alpha {
			alpha = score
			t.assignPV(height, move)
			if alpha >= beta {
				break
			}
		}
	}

	if !hasLegalMove {
		if !isCheck && skipMove == 0 {
			return valueDraw
		}
		return lossIn(height)
	}

	if alpha > oldAlpha && bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		historyContext.Update(quietsSearched, bestMove, depth)
		t.updateKiller(bestMove, height)
	}

	if skipMove == 0 {
		ttBound = 0
		if best > oldAlpha {
			ttBound |= boundLower
		}
		if best < beta {
			ttBound |= boundUpper
		}
		if !(rootNode && ttBound == boundUpper) {
			t.sharedContext.transTable.Update(position.Key, depth, valueToTT(best, height), ttBound, bestMove)
		}
	}

	return best
}

func (t *thread) quiescence(alpha, beta, height int) int {
	t.clearPV(height)
	var position = &t.stack[height].position
	if isDraw(position) {
		return valueDraw
	}
	if height >= maxHeight {
		return t.evaluator.EvaluateQuick(position)
	}
	if t.isRepeat(height) {
		return valueDraw
	}

	var _, ttValue, ttBound, _, ttHit = t.sharedContext.transTable.Read(position.Key)
	if ttHit {
		ttValue = valueFromTT(ttValue, height)
		if ttBound == boundExact ||
			ttBound == boundLower && ttValue >= beta ||
			ttBound == boundUpper && ttValue <= alpha {
			return ttValue
		}
	}

	var isCheck = position.IsCheck()
	var best = -valueInfinity
	if !isCheck {
		var eval = t.evaluator.EvaluateQuick(position)
		best = Max(best, eval)
		if eval > alpha {
			alpha = eval
			if alpha >= beta {
				return alpha
			}
		}
	}
	var mi = moveIteratorQS{
		position: position,
		buffer:   t.stack[height].moveList[:],
	}
	mi.Init()
	var hasLegalMove = false
	for mi.Reset(); ; {
		var move = mi.Next()
		if move == MoveEmpty {
			break
		}
		if !isCheck && !seeGEZero(position, move) {
			continue
		}
		if !t.makeMove(move, height) {
			continue
		}
		hasLegalMove = true
		var score = -t.quiescence(-beta, -alpha, height+1)
		t.unmakeMove()
		best = Max(best, score)
		if score > alpha {
			alpha = score
			t.assignPV(height, move)
			if alpha >= beta {
				break
			}
		}
	}
	if isCheck && !hasLegalMove {
		return lossIn(height)
	}
	return best
}

func (t *thread) makeMove(move Move, height int) bool {
	var pos = &t.stack[height].position
	var child = &t.stack[height+1].position
	if move == MoveEmpty {
		pos.MakeNullMove(child)
	} else {
		if !pos.MakeMove(move, child) {
			return false
		}
	}
	t.evaluator.MakeMove(pos, move)
	t.incNodes()
	return true
}

func (t *thread) unmakeMove() {
	t.evaluator.UnmakeMove()
}

func (t *thread) incNodes() {
	const (
		BATCH_SIZE = 1 << 8
		BATCH_MASK = BATCH_SIZE - 1
	)
	t.nodes += 1
	if t.nodes&BATCH_MASK == BATCH_MASK {
		t.sharedContext.AddNodes(BATCH_SIZE)
	}
}

func isDraw(p *Position) bool {
	if p.Rule50 > 100 {
		return true
	}

	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	return false
}

func (t *thread) isRepeat(height int) bool {
	var p = &t.stack[height].position

	if p.Rule50 == 0 || p.LastMove == MoveEmpty {
		return false
	}
	for i := height - 1; i >= 0; i-- {
		var temp = &t.stack[i].position
		if temp.Key == p.Key {
			return true
		}
		if temp.Rule50 == 0 || temp.LastMove == MoveEmpty {
			return false
		}
	}

	return t.sharedContext.game.IsTwoTimeRepeats(p.Key)
}

func (t *thread) updateKiller(move Move, height int) {
	if t.stack[height].killer1 != move {
		t.stack[height].killer2 = t.stack[height].killer1
		t.stack[height].killer1 = move
	}
}

func (t *thread) clearPV(height int) {
	t.stack[height].pv.size = 0
}

func (t *thread) assignPV(height int, m Move) {
	var pv = &t.stack[height].pv
	var child = &t.stack[height+1].pv
	pv.size = 1
	pv.items[0] = m
	if child.size > 0 {
		pv.size += child.size
		copy(pv.items[1:], child.items[:child.size])
	}
}

func (pv *pv) clone() []Move {
	var result = make([]Move, pv.size)
	copy(result, pv.items[:pv.size])
	return result
}

var reductions [64][64]int = initLmr(LmrMult)

func initLmr(f func(d, m float64) float64) [64][64]int {
	var res [64][64]int
	for d := 1; d < 64; d++ {
		for m := 1; m < 64; m++ {
			var r = f(float64(d), float64(m))
			res[d][m] = int(r)
		}
	}
	return res
}

func lmr(d, m int) int {
	return reductions[min(d, 63)][min(m, 63)]
}
