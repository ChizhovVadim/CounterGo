package engine

import (
	"runtime"
)

type UciOption interface {
	Name() string
}

type BoolUciOption struct {
	name  string
	Value bool
}

func (o *BoolUciOption) Name() string {
	return o.name
}

type IntUciOption struct {
	name            string
	Value, Min, Max int
}

func (o *IntUciOption) Name() string {
	return o.name
}

type Engine struct {
	Hash               IntUciOption
	Threads            IntUciOption
	ExperimentSettings BoolUciOption
	ClearTransTable    bool
	historyTable       historyTable
	transTable         *transTable
	evaluate           evaluate
	historyKeys        []uint64
	timeManager        *timeManager
	tree               [][]searchContext
}

func NewEngine() *Engine {
	var numCPUs = runtime.NumCPU()
	return &Engine{
		Hash:               IntUciOption{"Hash", 4, 4, 512},
		Threads:            IntUciOption{"Threads", numCPUs, 1, numCPUs},
		ExperimentSettings: BoolUciOption{"ExperimentSettings", false},
		historyTable:       NewHistoryTable(),
	}
}

func (e *Engine) GetInfo() (name, version, author string) {
	return "Counter", "2.1.0", "Vadim Chizhov"
}

func (e *Engine) GetOptions() []UciOption {
	return []UciOption{
		&e.Hash, &e.Threads, &e.ExperimentSettings}
}

func (e *Engine) Prepare() {
	if e.transTable == nil || e.transTable.megabytes != e.Hash.Value {
		e.transTable = NewTransTable(e.Hash.Value)
	}
	if len(e.tree) != e.Threads.Value {
		e.tree = NewTree(e, e.Threads.Value)
	}
}

func (e *Engine) Search(searchParams SearchParams) SearchInfo {
	var p = searchParams.Positions[len(searchParams.Positions)-1]
	e.timeManager = NewTimeManager(searchParams.Limits, TimeControlBasic,
		p.WhiteMove, searchParams.CancellationToken)
	defer e.timeManager.Close()

	e.Prepare()
	e.evaluate = Evaluate
	e.clearKillers()
	e.historyTable.Clear()
	e.transTable.PrepareNewSearch()
	if e.ClearTransTable {
		e.transTable.Clear()
	}
	e.historyKeys = PositionsToHistoryKeys(searchParams.Positions)
	for i := 0; i < len(e.tree); i++ {
		e.tree[i][0].Position = p
	}
	var ctx = &e.tree[0][0]
	return ctx.IterateSearch(searchParams.Progress)
}

func (e *Engine) clearKillers() {
	for i := 0; i < len(e.tree); i++ {
		for j := 0; j < len(e.tree[i]); j++ {
			var ctx = &e.tree[i][j]
			ctx.Killer1 = MoveEmpty
			ctx.Killer2 = MoveEmpty
		}
	}
}

func PositionsToHistoryKeys(positions []*Position) []uint64 {
	var result []uint64
	for _, p := range positions {
		if p.Rule50 == 0 {
			result = result[:0]
		}
		result = append(result, p.Key)
	}
	return result
}

func NewTree(engine *Engine, degreeOfParallelism int) [][]searchContext {
	var result = make([][]searchContext, degreeOfParallelism)
	for i := range result {
		result[i] = NewSearchContexts(engine, i)
	}
	return result
}

func NewSearchContexts(engine *Engine, thread int) []searchContext {
	var result = make([]searchContext, MAX_HEIGHT+1)
	for i := range result {
		result[i] = NewSearchContext(engine, thread, i)
	}
	return result
}

func NewSearchContext(engine *Engine, thread, height int) searchContext {
	return searchContext{
		Engine:             engine,
		Thread:             thread,
		Height:             height,
		Position:           &Position{},
		MoveList:           &MoveList{},
		QuietsSearched:     make([]Move, 0, MAX_MOVES),
		PrincipalVariation: make([]Move, 0, MAX_HEIGHT),
	}
}
