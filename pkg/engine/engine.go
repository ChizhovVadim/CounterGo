package engine

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Engine struct {
	Config     Config
	transTable TransTable
	threads    []thread
}

type Config struct {
	Hash               int
	Threads            int
	ExperimentSettings bool
	ProgressMinNodes   int
	EvalBuilder        func() IUpdatableEvaluator
}

func NewConfig(evalBuilder func() IUpdatableEvaluator) Config {
	return Config{
		Hash:               16,
		Threads:            1,
		ExperimentSettings: false,
		ProgressMinNodes:   1_000_000,
		EvalBuilder:        evalBuilder,
	}
}

type SharedContext struct {
	done        <-chan struct{}
	progress    func(common.SearchInfo)
	config      Config
	transTable  TransTable
	position    common.Position
	historyKeys map[uint64]int
	timeManager TimeManager
	start       time.Time
	nodes       atomic.Int64
}

func (c *SharedContext) isDone() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *SharedContext) ToSearchInfo(result mainLine) common.SearchInfo {
	return common.SearchInfo{
		Depth:    result.depth,
		MainLine: result.moves,
		Score:    newUciScore(result.score),
		Nodes:    c.nodes.Load(),
		Time:     time.Since(c.start),
	}
}

func (c *SharedContext) AddNodes(delta int) {
	var nodes = c.nodes.Add(int64(delta))
	if c.isDone() ||
		c.timeManager.IsFixedNodesInterruption(nodes) {
		panic(errSearchTimeout)
	}
}

func (c *SharedContext) OnIterationComplete(result mainLine) {
	if c.isDone() ||
		c.timeManager.IsSoftInterruption(result) {
		panic(errSearchTimeout)
	}
	if c.progress != nil &&
		c.nodes.Load() >= int64(c.config.ProgressMinNodes) {
		c.progress(c.ToSearchInfo(result))
	}
}

type mainLine struct {
	moves []common.Move
	score int
	depth int
}

type TimeManager interface {
	HardLimit() time.Duration
	IsFixedNodesInterruption(nodes int64) bool
	IsSoftInterruption(line mainLine) bool
}

type IUpdatableEvaluator interface {
	Init(p *common.Position)
	MakeMove(p *common.Position, m common.Move)
	UnmakeMove()
	EvaluateQuick(p *common.Position) int
}

type TransTable interface {
	Size() (megabytes int)
	IncDate()
	Clear()
	Read(key uint64) (depth, score, bound int, move common.Move, found bool)
	Update(key uint64, depth, score, bound int, move common.Move)
}

func New(config Config) *Engine {
	return &Engine{Config: config}
}

func (e *Engine) Clear() {
	if e.transTable != nil {
		e.transTable.Clear()
	}
	for i := range e.threads {
		e.threads[i].Clear()
	}
}

func (e *Engine) Prepare() {
	if e.transTable == nil || e.transTable.Size() != e.Config.Hash {
		if e.transTable != nil {
			// GC can collect TT
			// warning: thread.sharedContext keeps transTable
			// better transTable.Resize?
			e.transTable = nil
		}
		e.transTable = newTransTable(e.Config.Hash)
	}
	if len(e.threads) != e.Config.Threads {
		e.threads = make([]thread, e.Config.Threads)
		for i := range e.threads {
			var t = &e.threads[i]
			t.idx = i
			t.evaluator = e.Config.EvalBuilder()
		}
	}
}

func (e *Engine) Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo {
	var start = time.Now()
	e.Prepare()
	var p = &searchParams.Positions[len(searchParams.Positions)-1]
	var timeManager = newTimeManager(start, searchParams.Limits, p)

	var cancel context.CancelFunc
	if hardLimit := timeManager.HardLimit(); hardLimit > 0 {
		ctx, cancel = context.WithDeadline(ctx, start.Add(hardLimit))
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	e.transTable.IncDate()
	var historyKeys = getHistoryKeys(searchParams.Positions)

	var sharedContext = &SharedContext{
		done:        ctx.Done(),
		progress:    searchParams.Progress,
		config:      e.Config,
		transTable:  e.transTable,
		position:    *p,
		historyKeys: historyKeys,
		timeManager: timeManager,
		start:       start,
	}

	if e.Config.Threads == 1 {
		var mainLine = e.threads[0].iterativeDeepening(sharedContext)
		return sharedContext.ToSearchInfo(mainLine)
	} else {
		var mainLine = e.lazySmp(ctx, sharedContext)
		return sharedContext.ToSearchInfo(mainLine)
	}
}

func getHistoryKeys(positions []common.Position) map[uint64]int {
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
