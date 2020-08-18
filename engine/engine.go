package engine

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/ChizhovVadim/CounterGo/common"
)

type Engine struct {
	Hash               int
	Threads            int
	ExperimentSettings bool
	evalBuilder        func() Evaluator
	timeManager        TimeManager
	transTable         TransTable
	lateMoveReduction  func(d, m int) int
	historyKeys        map[uint64]int
	done               <-chan struct{}
	threads            []thread
	progress           func(SearchInfo)
	mainLine           mainLine
	start              time.Time
	nodes              int64
	depth              int32 // duplication mainLine.depth
	mu                 sync.Mutex
}

type thread struct {
	engine    *Engine
	sortTable SortTable
	evaluator Evaluator
	nodes     int64
	depth     int32
	stack     [stackSize]struct {
		position       Position
		moveList       [MaxMoves]OrderedMove
		quietsSearched [MaxMoves]Move
		pv             pv
		staticEval     int
	}
}

type pv struct {
	items [stackSize]Move
	size  int
}

type mainLine struct {
	moves []Move
	score int
	depth int
}

type TimeManager interface {
	OnNodesChanged(nodes int)
	OnIterationComplete(line mainLine)
	Close()
}

type Evaluator interface {
	Evaluate(p *Position) int
}

type SortTable interface {
	Clear()
	Update(p *Position, bestMove Move, searched []Move, depth, height int)
	Note(p *Position, ml []OrderedMove, trans Move, height int)
	NoteQS(p *Position, ml []OrderedMove)
}

type TransTable interface {
	Megabytes() int
	PrepareNewSearch()
	Clear()
	Read(p *Position) (depth, score, bound int, move Move, ok bool)
	Update(p *Position, depth, score, bound int, move Move)
}

func NewEngine(evalBuilder func() Evaluator) *Engine {
	return &Engine{
		Hash:               16,
		Threads:            1,
		ExperimentSettings: false,
		evalBuilder:        evalBuilder,
	}
}

func (e *Engine) Prepare() {
	if e.transTable == nil || e.transTable.Megabytes() != e.Hash {
		if e.transTable != nil {
			e.transTable = nil
			runtime.GC()
		}
		e.transTable = newTransTable(e.Hash)
	}
	if e.lateMoveReduction == nil {
		e.lateMoveReduction = initLmr(lmrMult)
	}
	if len(e.threads) != e.Threads {
		e.threads = make([]thread, e.Threads)
		for i := range e.threads {
			var t = &e.threads[i]
			t.engine = e
			t.sortTable = &sortTable{}
			t.evaluator = e.evalBuilder()
		}
	}
}

func (e *Engine) Search(ctx context.Context, searchParams SearchParams) SearchInfo {
	e.start = time.Now()
	e.Prepare()
	var p = &searchParams.Positions[len(searchParams.Positions)-1]
	ctx, e.timeManager = withTimeManager(ctx, e.start, searchParams.Limits, p)
	defer e.timeManager.Close()
	e.transTable.PrepareNewSearch()
	e.historyKeys = getHistoryKeys(searchParams.Positions)
	e.nodes = 0
	for i := range e.threads {
		var t = &e.threads[i]
		t.nodes = 0
		t.stack[0].position = *p
	}
	e.progress = searchParams.Progress
	lazySmp(ctx, e)
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
		t.sortTable.Clear()
	}
}

func (e *Engine) currentSearchResult() SearchInfo {
	return SearchInfo{
		Depth:    e.mainLine.depth,
		MainLine: e.mainLine.moves,
		Score:    newUciScore(e.mainLine.score),
		Nodes:    atomic.LoadInt64(&e.nodes),
		Time:     int64(time.Since(e.start) / time.Millisecond),
	}
}

func (e *Engine) sendProgress() {
	if e.progress != nil {
		e.progress(e.currentSearchResult())
	}
}

func (pv *pv) clear() {
	pv.size = 0
}

func (pv *pv) assign(m Move, child *pv) {
	pv.size = 1 + child.size
	pv.items[0] = m
	copy(pv.items[1:], child.items[:child.size])
}

func (pv *pv) toSlice() []Move {
	var result = make([]Move, pv.size)
	copy(result, pv.items[:pv.size])
	return result
}
