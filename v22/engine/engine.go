package engine

import (
	"runtime"

	. "github.com/ChizhovVadim/CounterGo/common"
)

type Engine struct {
	Hash               IntUciOption
	Threads            IntUciOption
	ExperimentSettings BoolUciOption
	ClearTransTable    bool
	historyTable       historyTable
	transTable         *transTable
	evaluate           evaluate
	historyKeys       map[uint64]int
	timeManager        *timeManager
	tree               [][]searchContext
}

type searchContext struct {
	Engine             *Engine
	Thread             int
	Height             int
	Position           *Position
	MoveList           *MoveList
	mi                 moveIterator
	Killer1            Move
	Killer2            Move
	PrincipalVariation []Move
	QuietsSearched     []Move
}

type evaluate func(p *Position) int

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
	return "Counter", "2.2.0dev", "Vadim Chizhov"
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
	e.timeManager = NewTimeManager(searchParams.Limits, timeControlSmart,
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
	e.historyKeys = getHistoryKeys(searchParams.Positions)
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

func getHistoryKeys(positions []*Position) map[uint64]int {
	var result = make(map[uint64]int)
	for i := len(positions) - 1; i >= 0; i-- {
		var p = positions[i]
		if p.Rule50 == 0 {
			break
		}
		result[p.Key]++
	}
	return result
}

func NewTree(engine *Engine, degreeOfParallelism int) [][]searchContext {
	var result = make([][]searchContext, degreeOfParallelism)
	for i := 0; i < len(result); i++ {
		result[i] = NewSearchContexts(engine, i)
	}
	return result
}

func NewSearchContexts(engine *Engine, thread int) []searchContext {
	var result = make([]searchContext, MAX_HEIGHT+1)
	for i := 0; i < len(result); i++ {
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
