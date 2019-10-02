package engine

import (
	"context"
	"errors"
	"math/rand"
	"sync"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const pawnValue = 100

var errSearchTimeout = errors.New("search timeout")

func iterativeDeepening(ctx context.Context, e *Engine) {
	var ml = e.genRootMoves()
	if len(ml) != 0 {
		e.mainLine = mainLine{
			depth: 0,
			score: 0,
			moves: []Move{ml[0]},
		}
	}
	if len(ml) <= 1 {
		return
	}

	defer recoverFromSearchTimeout()
	e.done = ctx.Done()

	for depth := 1; depth <= maxHeight; depth++ {
		searchRoot(&e.threads[0], ml, depth, &e.mainLine)
		if e.timeManager.BreakIterativeDeepening(e.mainLine) {
			break
		}
		e.sendProgress()
		savePV(e.transTable, &e.threads[0].stack[0].position, e.mainLine.moves)
	}
}

func savePV(transTable TransTable, p *Position, pv []Move) {
	var parent = *p
	var child Position
	for _, m := range pv {
		transTable.Update(&parent, 0, 0, 0, m)
		parent.MakeMove(m, &child)
		parent = child
	}
}

func iterativeDeepeningLazySmp(ctx context.Context, e *Engine) {
	var ml = e.genRootMoves()
	if len(ml) != 0 {
		e.mainLine = mainLine{
			depth: 0,
			score: 0,
			moves: []Move{ml[0]},
		}
	}
	if len(ml) <= 1 {
		return
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	e.done = ctx.Done()

	var depths = make(chan int)
	var taskResults = make(chan mainLine)
	var wg = &sync.WaitGroup{}

	for i := 0; i < e.Threads.Value; i++ {
		var ml = cloneMoves(ml)
		wg.Add(1)
		go func(i int) {
			var t = &e.threads[i]
			if i > 0 {
				shuffleMoves(ml[1:])
			}
			for depth := range depths {
				func() {
					defer recoverFromSearchTimeout()
					var line mainLine
					searchRoot(t, ml, depth, &line)
					taskResults <- line
				}()
			}
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(taskResults)
	}()

	go func() {
		defer close(depths)
		for depth := 1; depth <= maxHeight; depth++ {
			var numThreads = threadsPerDepth(depth, e.Threads.Value)
			for i := 0; i < numThreads; i++ {
				select {
				case <-ctx.Done():
					return
				case depths <- depth:
				}
			}
		}
	}()

	for taskResult := range taskResults {
		if taskResult.depth > e.mainLine.depth {
			e.mainLine = taskResult
			if e.timeManager.BreakIterativeDeepening(e.mainLine) {
				cancel()
			}
			e.sendProgress()
		}
	}
}

func threadsPerDepth(depth, threads int) int {
	if depth <= 4 {
		return 1
	}
	return 1 + (threads-1)/2
}

func searchRoot(t *thread, ml []Move, depth int, line *mainLine) {
	const height = 0
	t.stack[height].pv.clear()
	t.stack[height].pvNode = true
	var p = &t.stack[height].position
	var child = &t.stack[height+1].position
	var alpha = -valueInfinity
	const beta = valueInfinity
	var bestMoveIndex = 0
	for i, move := range ml {
		p.MakeMove(move, child)
		var newDepth = t.newDepth(depth, height)
		var reduction = 0
		if depth >= 3 && !(i == 0 ||
			p.IsCheck() ||
			child.IsCheck() ||
			isCaptureOrPromotion(move) ||
			isPawnAdvance(move, p.WhiteMove)) {
			reduction = t.engine.lateMoveReduction(depth, i+1)
			reduction--
			reduction = Max(0, Min(depth-2, reduction))
		}
		if reduction > 0 &&
			-t.alphaBeta(-(alpha+1), -alpha, newDepth-reduction, height+1) <= alpha {
			continue
		}
		if beta != alpha+1 && i > 0 && newDepth > 0 &&
			-t.alphaBeta(-(alpha+1), -alpha, newDepth, height+1) <= alpha {
			continue
		}
		var score = -t.alphaBeta(-beta, -alpha, newDepth, height+1)
		if score > alpha {
			alpha = score
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
			*line = mainLine{
				depth: depth,
				score: score,
				moves: t.stack[height].pv.toSlice(),
			}
			bestMoveIndex = i
		}
	}
	moveToBegin(ml, bestMoveIndex)
}

func (t *thread) alphaBeta(alpha, beta, depth, height int) int {
	var newDepth, score int
	t.stack[height].pv.clear()

	var position = &t.stack[height].position

	if height >= maxHeight {
		return t.evaluator.Evaluate(position)
	}

	if t.isDraw(height) {
		return valueDraw
	}

	if depth <= 0 {
		return t.quiescence(alpha, beta, 1, height)
	}

	t.incNodes()

	var isCheck = position.IsCheck()

	// mate distance pruning
	if winIn(height+1) <= alpha {
		return alpha
	}
	if lossIn(height+2) >= beta && !isCheck {
		return beta
	}

	var pvNode = beta != alpha+1
	t.stack[height].pvNode = pvNode

	// transposition table
	var ttDepth, ttValue, ttBound, ttMove, ttHit = t.engine.transTable.Read(position)
	if ttHit {
		ttValue = valueFromTT(ttValue, height)
		if ttDepth >= depth && !pvNode {
			if ttValue >= beta && (ttBound&boundLower) != 0 {
				if ttMove != MoveEmpty && !isCaptureOrPromotion(ttMove) {
					t.sortTable.Update(position, ttMove, nil, depth, height)
				}
				return beta
			}
			if ttValue <= alpha && (ttBound&boundUpper) != 0 {
				return alpha
			}
		}
	}

	var lazyEval = lazyEval{evaluator: t.evaluator, position: position}
	var lateEndgame = isLateEndgame(position, position.WhiteMove)

	// reverse futility pruning
	if depth <= 2 && height >= 1 && !t.stack[height-1].pvNode && !isCheck &&
		beta < valueWin && beta > valueLoss &&
		lazyEval.Value()-pawnValue*depth >= beta {
		return beta
	}

	// null-move pruning
	var child = &t.stack[height+1].position
	if depth >= 2 && !pvNode && !isCheck && position.LastMove != MoveEmpty &&
		beta < valueWin && beta > valueLoss &&
		!(ttHit && ttValue < beta && (ttBound&boundUpper) != 0) &&
		!lateEndgame {
		newDepth = depth - 4
		position.MakeNullMove(child)
		// take eval with opponent side-to-move bonus(es)
		if -t.evaluator.Evaluate(child) >= beta {
			if newDepth <= 0 {
				score = -t.quiescence(-beta, -(beta - 1), 1, height+1)
			} else {
				score = -t.alphaBeta(-beta, -(beta - 1), newDepth, height+1)
			}
			if score >= beta {
				return beta
			}
		}
	}

	var ml = position.GenerateMoves(t.stack[height].moveList[:])
	t.sortTable.Note(position, ml, ttMove, height)

	// singular extension
	var ttMoveIsSingular = false
	if pvNode && depth >= 6 &&
		ttMove != MoveEmpty && ttDepth >= depth-4 && (ttBound&boundLower) != 0 &&
		ttValue < valueWin && ttValue > valueLoss {
		ttMoveIsSingular = true
		sortMoves(ml)
		var bound = Max(-valueInfinity, ttValue-pawnValue/2)
		newDepth = depth - 5
		for i := range ml {
			var move = ml[i].Move
			if !position.MakeMove(move, child) {
				continue
			}
			if move == ttMove {
				if t.newDepth(depth, height) == depth {
					break
				}
				continue
			}
			score = -t.alphaBeta(-(bound + 1), -bound, newDepth, height+1)
			if score > bound {
				ttMoveIsSingular = false
				break
			}
		}
	}

	var moveCount = 0
	var quietsSearched = t.stack[height].quietsSearched[:0]
	var bestMove Move
	const SortMovesIndex = 4

	for i := range ml {
		if i < SortMovesIndex {
			moveToTop(ml[i:])
		} else if i == SortMovesIndex {
			sortMoves(ml[i:])
		}
		var move = ml[i].Move

		if !position.MakeMove(move, child) {
			continue
		}
		moveCount++

		newDepth = t.newDepth(depth, height)
		if move == ttMove && ttMoveIsSingular {
			newDepth = depth
		}

		var reduction = 0
		if !(moveCount == 1 ||
			ml[i].Key >= sortTableKeyImportant ||
			isCheck ||
			child.IsCheck() ||
			isCaptureOrPromotion(move) ||
			isPawnAdvance(move, position.WhiteMove)) {

			if depth >= 3 {
				reduction = t.engine.lateMoveReduction(depth, moveCount)
				if pvNode {
					reduction--
				}
				reduction = Max(0, Min(depth-2, reduction))
			}

			if !(pvNode ||
				alpha <= valueLoss ||
				position.LastMove == MoveEmpty) {
				// late-move pruning
				if depth <= 2 && moveCount >= 9+3*depth {
					continue
				}
				// futility pruning
				if depth <= 3 && lazyEval.Value()+pawnValue*depth <= alpha {
					continue
				}
			}
		}

		if !isCaptureOrPromotion(move) {
			quietsSearched = append(quietsSearched, move)
		}

		// LMR
		if reduction > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth-reduction, height+1)
			if score <= alpha {
				continue
			}
		}

		// PVS
		if beta != alpha+1 && moveCount > 1 && newDepth > 0 {
			score = -t.alphaBeta(-(alpha + 1), -alpha, newDepth, height+1)
			if score <= alpha {
				continue
			}
		}

		score = -t.alphaBeta(-beta, -alpha, newDepth, height+1)

		if score > alpha {
			alpha = score
			bestMove = move
			if alpha >= beta {
				break
			}
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
		}
	}

	if moveCount == 0 {
		if isCheck {
			return lossIn(height)
		}
		return valueDraw
	}

	if bestMove != MoveEmpty && !isCaptureOrPromotion(bestMove) {
		t.sortTable.Update(position, bestMove, quietsSearched, depth, height)
	}

	ttBound = 0
	if bestMove != MoveEmpty {
		ttBound |= boundLower
	}
	if alpha < beta {
		ttBound |= boundUpper
	}
	t.engine.transTable.Update(position, depth, valueToTT(alpha, height), ttBound, bestMove)

	return alpha
}

func (t *thread) quiescence(alpha, beta, depth, height int) int {
	t.stack[height].pv.clear()
	t.incNodes()
	var position = &t.stack[height].position
	if height >= maxHeight {
		return t.evaluator.Evaluate(position)
	}
	var isCheck = position.IsCheck()
	var eval = 0
	if !isCheck {
		eval = t.evaluator.Evaluate(position)
		if eval > alpha {
			alpha = eval
		}
		if eval >= beta {
			return alpha
		}
	}
	var ml = t.stack[height].moveList[:]
	if position.IsCheck() {
		ml = position.GenerateMoves(ml)
	} else {
		ml = position.GenerateCaptures(ml, depth > 0)
	}
	t.sortTable.NoteQS(position, ml)
	sortMoves(ml)
	var moveCount = 0
	var child = &t.stack[height+1].position
	for i := range ml {
		var move = ml[i].Move
		var danger = isDangerCapture(position, move)
		if !isCheck && !danger && !seeGEZero(position, move) {
			continue
		}
		if !position.MakeMove(move, child) {
			continue
		}
		moveCount++
		if !isCheck && !danger && !child.IsCheck() &&
			eval+moveValue(move)+2*pawnValue <= alpha {
			continue
		}
		var score = -t.quiescence(-beta, -alpha, depth-1, height+1)
		if score > alpha {
			alpha = score
			if score >= beta {
				break
			}
			t.stack[height].pv.assign(move, &t.stack[height+1].pv)
		}
	}
	if isCheck && moveCount == 0 {
		return lossIn(height)
	}
	return alpha
}

func (t *thread) incNodes() {
	t.nodes++
	if (t.nodes&255) == 0 && isDone(t.engine.done) {
		panic(errSearchTimeout)
	}
}

func isDone(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

func (t *thread) isDraw(height int) bool {
	var p = &t.stack[height].position

	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	if p.Rule50 > 100 {
		return true
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

	if t.engine.historyKeys[p.Key] >= 2 {
		return true
	}

	return false
}

/*func (t *thread) newDepth(depth, height int) int {
	var p = &t.stack[height].position
	var child = &t.stack[height+1].position
	var move = child.LastMove
	var givesCheck = child.IsCheck()

	if givesCheck && (depth <= 1 || seeGEZero(p, move)) {
		return depth
	}

	var prevMove = p.LastMove

	if prevMove != MoveEmpty &&
		prevMove.To() == move.To() &&
		move.CapturedPiece() > Pawn &&
		prevMove.CapturedPiece() > Pawn &&
		seeGEZero(p, move) {
		return depth
	}

	if isPawnPush7th(move, p.WhiteMove) && seeGEZero(p, move) {
		return depth
	}

	return depth - 1
}*/

func (t *thread) newDepth(depth, height int) int {
	return depth - 1
}

func recoverFromSearchTimeout() {
	var r = recover()
	if r != nil && r != errSearchTimeout {
		panic(r)
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

func moveToTop(ml []OrderedMove) {
	var bestIndex = 0
	for i := 1; i < len(ml); i++ {
		if ml[i].Key > ml[bestIndex].Key {
			bestIndex = i
		}
	}
	if bestIndex != 0 {
		ml[0], ml[bestIndex] = ml[bestIndex], ml[0]
	}
}

func cloneMoves(ml []Move) []Move {
	var result = make([]Move, len(ml))
	copy(result, ml)
	return result
}

func shuffleMoves(ml []Move) {
	rand.Shuffle(len(ml), func(i, j int) {
		ml[i], ml[j] = ml[j], ml[i]
	})
}

func (e *Engine) genRootMoves() []Move {
	var t = e.threads[0]
	const height = 0
	var p = &t.stack[height].position
	_, _, _, transMove, _ := e.transTable.Read(p)
	var ml = p.GenerateMoves(t.stack[height].moveList[:])
	t.sortTable.Note(p, ml, transMove, height)
	sortMoves(ml)
	var result []Move
	var child = &t.stack[height+1].position
	for i := range ml {
		var move = ml[i].Move
		if p.MakeMove(move, child) {
			result = append(result, move)
		}
	}
	return result
}

type lazyEval struct {
	evaluator Evaluator
	position  *Position
	hasValue  bool
	value     int
}

func (le *lazyEval) Value() int {
	if !le.hasValue {
		le.value = le.evaluator.Evaluate(le.position)
		le.hasValue = true
	}
	return le.value
}
