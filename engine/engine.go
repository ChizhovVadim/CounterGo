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

type Evaluator interface {
	Evaluate(p *Position) int
	MoveValue(move Move) int
}

type Engine struct {
	Hash               IntUciOption
	Threads            IntUciOption
	ExperimentSettings BoolUciOption
	ClearTransTable    bool
	historyTable       historyTable
	transTable         *transTable
	evaluator          Evaluator
	historyKeys        []uint64
	timeManager        *timeManager
	tree               [][]searchContext
}

type searchContext struct {
	Engine             *Engine
	Thread             int
	Height             int
	Position           *Position
	mi                 moveIterator
	Killer1            Move
	Killer2            Move
	PrincipalVariation []Move
	QuietsSearched     []Move
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
		e.initTree()
	}
	if e.evaluator == nil {
		e.evaluator = NewEvaluation(e.ExperimentSettings.Value)
	}
}

func (e *Engine) Search(searchParams SearchParams) SearchInfo {
	var p = searchParams.Positions[len(searchParams.Positions)-1]
	e.timeManager = NewTimeManager(searchParams.Limits, timeControlSmart,
		p.WhiteMove, searchParams.Ctx)
	defer e.timeManager.Close()

	e.Prepare()
	e.clearKillers()
	e.historyTable.Clear()
	e.transTable.PrepareNewSearch()
	if e.ClearTransTable {
		e.transTable.Clear()
	}
	e.historyKeys = PositionsToHistoryKeys(searchParams.Positions)
	for thread := range e.tree {
		e.tree[thread][0].Position = p
	}
	var ctx = &e.tree[0][0]
	return ctx.IterateSearch(searchParams.Progress)
}

func (e *Engine) clearKillers() {
	for thread := range e.tree {
		for height := range e.tree[thread] {
			var ctx = &e.tree[thread][height]
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

func (e *Engine) initTree() {
	e.tree = make([][]searchContext, e.Threads.Value)
	for thread := range e.tree {
		e.tree[thread] = make([]searchContext, MAX_HEIGHT+1)
		for height := range e.tree[thread] {
			e.tree[thread][height] = searchContext{
				Engine:             e,
				Thread:             thread,
				Height:             height,
				Position:           &Position{},
				QuietsSearched:     make([]Move, 0, MAX_MOVES),
				PrincipalVariation: make([]Move, 0, MAX_HEIGHT),
			}
		}
	}
}
