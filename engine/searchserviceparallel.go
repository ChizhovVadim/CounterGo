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
			var score = -this.AlphaBeta(child, -beta, -alpha, newDepth, height+1, false)
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
					var score = -this.AlphaBeta(child, -(localAlpha + 1), -localAlpha, newDepth, height+1, true)
					if IsCancelValue(score) {
						return
					}
					if score <= localAlpha {
						continue
					}
				}
				var score = -this.AlphaBeta(child, -beta, -localAlpha, newDepth, height+1, false)
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
