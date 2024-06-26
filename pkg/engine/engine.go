package engine

import (
	"context"
	"errors"
	"time"

	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Engine struct {
	Options     Options
	timeManager TimeManager
	transTable  TransTable
	historyKeys map[uint64]int
	threads     []thread
	progress    func(SearchInfo)
	mainLine    mainLine
	start       time.Time
}

type thread struct {
	engine    *Engine
	evaluator IUpdatableEvaluator
	nodes     int64
	rootDepth int
	stack     [stackSize]struct {
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

type mainLine struct {
	moves []Move
	score int
	depth int
	nodes int64
}

type TimeManager interface {
	IsDone() bool
	OnNodesChanged(nodes int)
	OnIterationComplete(line mainLine)
	Close()
}

type IEvaluator interface {
	Evaluate(p *Position) int
}

type IUpdatableEvaluator interface {
	Init(p *Position)
	MakeMove(p *Position, m Move)
	UnmakeMove()
	EvaluateQuick(p *Position) int
}

type TransTable interface {
	Size() (megabytes int)
	IncDate()
	Clear()
	Read(key uint64) (depth, score, bound int, move Move, found bool)
	Update(key uint64, depth, score, bound int, move Move)
}

func NewEngine(options Options) *Engine {
	return &Engine{
		Options: options,
	}
}

func (e *Engine) Prepare() {
	if e.transTable == nil || e.transTable.Size() != e.Options.Hash {
		if e.transTable != nil {
			// GC can collect TT
			e.transTable = nil
		}
		e.transTable = newTransTable(e.Options.Hash)
	}
	if len(e.threads) != e.Options.Threads {
		e.threads = make([]thread, e.Options.Threads)
		for i := range e.threads {
			var t = &e.threads[i]
			t.engine = e
			t.evaluator = e.buildEvaluator()
		}
	}
}

func (e *Engine) Search(ctx context.Context, searchParams SearchParams) SearchInfo {
	e.start = time.Now()
	e.Prepare()
	var p = &searchParams.Positions[len(searchParams.Positions)-1]
	e.timeManager = newTimeManager(ctx, e.start, searchParams.Limits, p)
	defer e.timeManager.Close()
	e.transTable.IncDate()
	e.historyKeys = getHistoryKeys(searchParams.Positions)
	e.mainLine.nodes = 0
	for i := range e.threads {
		var t = &e.threads[i]
		t.nodes = 0
		t.stack[0].position = *p
	}
	e.progress = searchParams.Progress
	lazySmp(e)
	for i := range e.threads {
		var t = &e.threads[i]
		e.mainLine.nodes += t.nodes
		t.nodes = 0
	}
	return e.currentSearchResult()
}

func getHistoryKeys(positions []Position) map[uint64]int {
	var result = make(map[uint64]int)
	for i := len(positions) - 1; i >= 0; i-- {
		var p = &positions[i]
		result[p.Key]++
		if p.Rule50 == 0 {
			break
		}
	}
	return result
}

func (e *Engine) Clear() {
	if e.transTable != nil {
		e.transTable.Clear()
	}
	for i := range e.threads {
		var t = &e.threads[i]
		t.clearHistory()
	}
}

func (e *Engine) currentSearchResult() SearchInfo {
	return SearchInfo{
		Depth:    e.mainLine.depth,
		MainLine: e.mainLine.moves,
		Score:    newUciScore(e.mainLine.score),
		Nodes:    e.mainLine.nodes,
		Time:     time.Since(e.start),
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

func (pv *pv) toSlice() []Move {
	var result = make([]Move, pv.size)
	copy(result, pv.items[:pv.size])
	return result
}

type EvaluatorAdapter struct {
	evaluator IEvaluator
}

func (e *EvaluatorAdapter) Init(p *Position) {

}

func (e *EvaluatorAdapter) MakeMove(p *Position, m Move) {

}

func (e *EvaluatorAdapter) UnmakeMove() {

}

func (e *EvaluatorAdapter) EvaluateQuick(p *Position) int {
	return e.evaluator.Evaluate(p)
}

func (e *Engine) buildEvaluator() IUpdatableEvaluator {
	var evaluationService = e.Options.EvalBuilder()
	if ue, ok := evaluationService.(IUpdatableEvaluator); ok {
		return ue
	}
	if e, ok := evaluationService.(IEvaluator); ok {
		return &EvaluatorAdapter{evaluator: e}
	}
	panic(errors.New("bad eval builder"))
}
