package engine

import (
	"context"
	"runtime"

	. "github.com/ChizhovVadim/CounterGo/common"
)

type Engine struct {
	Hash               IntUciOption
	Threads            IntUciOption
	ExperimentSettings BoolUciOption
	ClearTransTable    bool
	timeManager        *timeManager
	transTable         TransTable
	threads            []thread
}

type thread struct {
	sortTable         SortTable
	pvTable           *pvTable
	transTable        TransTable
	evaluator         Evaluator
	lateMoveReduction func(d, m int) int
	historyKeys       map[uint64]int
	done              <-chan struct{}
	nodes             int
	stack             [stackSize]struct {
		position       Position
		moveList       [MaxMoves]OrderedMove
		quietsSearched [MaxMoves]Move
	}
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

func NewEngine() *Engine {
	var numCPUs = runtime.NumCPU()
	return &Engine{
		Hash:               IntUciOption{Name: "Hash", Value: 4, Min: 4, Max: 512},
		Threads:            IntUciOption{Name: "Threads", Value: numCPUs, Min: 1, Max: numCPUs},
		ExperimentSettings: BoolUciOption{Name: "ExperimentSettings", Value: false},
	}
}

func (e *Engine) GetInfo() (name, version, author string) {
	return "Counter", "2.9dev", "Vadim Chizhov"
}

func (e *Engine) GetOptions() []UciOption {
	return []UciOption{
		&e.Hash, &e.Threads, &e.ExperimentSettings}
}

func (e *Engine) Prepare() {
	if e.transTable == nil || e.transTable.Megabytes() != e.Hash.Value {
		e.transTable = NewTransTable(e.Hash.Value)
	}
	if len(e.threads) != e.Threads.Value {
		e.threads = make([]thread, e.Threads.Value)
		for thread := range e.threads {
			var t = &e.threads[thread]
			t.sortTable = NewSortTable()
			t.pvTable = NewPvTable()
			if e.ExperimentSettings.Value {
				t.evaluator = NewExperimentEvaluationService()
			} else {
				t.evaluator = NewEvaluationService()
			}
			t.lateMoveReduction = lmrTwo
		}
	}
}

func (e *Engine) Search(ctx context.Context, searchParams SearchParams) SearchInfo {
	var p = &searchParams.Positions[len(searchParams.Positions)-1]
	e.timeManager = NewTimeManager(searchParams.Limits, timeControlSmart, p.WhiteMove)
	if e.timeManager.hardTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeManager.hardTime)
		defer cancel()
	}
	e.Prepare()
	e.transTable.PrepareNewSearch()
	if e.ClearTransTable {
		e.transTable.Clear()
	}
	var historyKeys = getHistoryKeys(searchParams.Positions)
	for thread := range e.threads {
		var t = &e.threads[thread]
		t.nodes = 0
		t.done = ctx.Done()
		t.stack[0].position = *p
		t.historyKeys = historyKeys
		t.transTable = e.transTable
		t.sortTable.Clear()
		t.pvTable.Clear()
	}
	return e.iterateSearch(searchParams.Progress)
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
