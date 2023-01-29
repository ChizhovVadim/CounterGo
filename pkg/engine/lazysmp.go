package engine

import (
	"errors"
	"sync"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

var errSearchTimeout = errors.New("search timeout")

type searchTask struct {
	depth         int
	startingMove  common.Move //for move ordering
	startingScore int         //for aspirationWindow
}

func lazySmp(e *Engine) {
	var ml = e.genRootMoves()
	if len(ml) != 0 {
		e.mainLine = mainLine{
			depth: 0,
			score: 0,
			nodes: 0,
			moves: []common.Move{ml[0]},
		}
	}
	if len(ml) <= 1 {
		return
	}

	var tasks = make(chan searchTask)
	var taskResults = make(chan mainLine)

	var wg = &sync.WaitGroup{}

	for i := 0; i < e.Threads; i++ {
		wg.Add(1)
		go func(t *thread, ml []common.Move) {
			defer wg.Done()
			searchDepth(t, ml, tasks, taskResults)
		}(&e.threads[i], cloneMoves(ml))
	}

	go func() {
		wg.Wait()
		close(taskResults)
	}()

	iterativeDeepening(e, tasks, taskResults)
}

func iterativeDeepening(
	e *Engine,
	tasks chan<- searchTask,
	taskResults <-chan mainLine,
) {
	var searchCountByDepth [stackSize]int
	for {
		var task = searchTask{
			depth:         e.mainLine.depth + 1, // next Iteration
			startingMove:  e.mainLine.moves[0],
			startingScore: e.mainLine.score,
		}
		if task.depth < len(searchCountByDepth) &&
			searchCountByDepth[task.depth] >= (e.Threads+1)/2 {
			// some threads search deeper
			task.depth = e.mainLine.depth + 2
		}

		if task.depth > maxHeight ||
			e.timeManager.IsDone() {
			// no new iterations
			if tasks != nil {
				close(tasks)
				tasks = nil
			}
		}

		select {
		case taskResult, ok := <-taskResults:
			if !ok {
				// all searches finished
				return
			}
			e.mainLine.nodes += taskResult.nodes
			if taskResult.depth > e.mainLine.depth {
				e.mainLine.depth = taskResult.depth
				e.mainLine.score = taskResult.score
				e.mainLine.moves = taskResult.moves
				e.timeManager.OnIterationComplete(e.mainLine)
				if e.progress != nil && e.mainLine.nodes >= int64(e.ProgressMinNodes) {
					e.progress(e.currentSearchResult())
				}
			}
		case tasks <- task:
			searchCountByDepth[task.depth]++
		}
	}
}

func searchDepth(
	t *thread,
	ml []common.Move,
	tasks <-chan searchTask,
	taskResults chan<- mainLine,
) {
	defer func() {
		if r := recover(); r != nil {
			if r == errSearchTimeout {
				return
			}
			panic(r)
		}
	}()

	const height = 0
	for h := 0; h <= 2; h++ {
		t.stack[h].killer1 = common.MoveEmpty
		t.stack[h].killer2 = common.MoveEmpty
	}

	for task := range tasks {
		if task.startingMove != common.MoveEmpty {
			var index = findMoveIndex(ml, task.startingMove)
			if index >= 0 {
				moveToBegin(ml, index)
			}
		}
		var score = aspirationWindow(t, ml, task.depth, task.startingScore)
		taskResults <- mainLine{
			depth: task.depth,
			score: score,
			moves: t.stack[height].pv.toSlice(),
			nodes: t.nodes,
		}
		t.nodes = 0
	}
}
