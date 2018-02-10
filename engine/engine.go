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
	transTable         TransTable
	evaluate           evaluate
	historyKeys        map[uint64]int
	timeManager        *timeManager
	tree               [][]node
	lateMoveReduction  func(d, m int) int
}

type node struct {
	engine             *Engine
	thread             int
	height             int
	position           *Position
	killer1            Move
	killer2            Move
	principalVariation []Move
	quietsSearched     []Move
	buffer0            [MaxMoves]Move
	buffer1            [MaxMoves]orderedMove
	buffer2            [MaxMoves]orderedMove
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
	return "Counter", "2.6dev", "Vadim Chizhov"
}

func (e *Engine) GetOptions() []UciOption {
	return []UciOption{
		&e.Hash, &e.Threads, &e.ExperimentSettings}
}

func (e *Engine) Prepare() {
	if e.transTable == nil || e.transTable.Megabytes() != e.Hash.Value {
		e.transTable = NewSimpleTransTable(e.Hash.Value)
	}
	if len(e.tree) != e.Threads.Value {
		e.initTree()
	}
	if e.lateMoveReduction == nil {
		e.lateMoveReduction = initLmrCrafty()
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
	for thread := range e.tree {
		e.tree[thread][0].position = p
	}
	var node = &e.tree[0][0]
	return node.IterateSearch(searchParams.Progress)
}

func (e *Engine) clearKillers() {
	for thread := range e.tree {
		for height := range e.tree[thread] {
			var node = &e.tree[thread][height]
			node.killer1 = MoveEmpty
			node.killer2 = MoveEmpty
		}
	}
}

func (e *Engine) initKillers() {
	for thread := 1; thread < len(e.tree); thread++ {
		for height := range e.tree[0] {
			e.tree[thread][height].killer1 = e.tree[0][height].killer1
			e.tree[thread][height].killer2 = e.tree[0][height].killer2
		}
	}
}

func getHistoryKeys(positions []*Position) map[uint64]int {
	var result = make(map[uint64]int)
	for i := len(positions) - 1; i >= 0; i-- {
		var p = positions[i]
		result[p.Key]++
		if p.Rule50 == 0 {
			break
		}
	}
	return result
}

func (e *Engine) initTree() {
	e.tree = make([][]node, e.Threads.Value)
	for thread := range e.tree {
		e.tree[thread] = make([]node, maxHeight+1)
		for height := range e.tree[thread] {
			e.tree[thread][height] = node{
				engine:             e,
				thread:             thread,
				height:             height,
				position:           &Position{},
				quietsSearched:     make([]Move, 0, MaxMoves),
				principalVariation: make([]Move, 0, maxHeight),
			}
		}
	}
}
