package engine

import (
	"errors"
	"sync/atomic"
	"time"

	. "github.com/ChizhovVadim/CounterGo/common"
)

var searchTimeout = errors.New("search timeout")

type timeControlStrategy func(main, inc, moves int) (softLimit, hardLimit int)

type timeManager struct {
	start    time.Time
	softTime time.Duration
	nodes    int64
	ct       *CancellationToken
	timer    *time.Timer
}

func (tm *timeManager) Nodes() int64 {
	return tm.nodes
}

func (tm *timeManager) IsHardTimeout() bool {
	return tm.ct.IsCancellationRequested()
}

func (tm *timeManager) IncNodes() {
	atomic.AddInt64(&tm.nodes, 1)
	if tm.IsHardTimeout() {
		panic(searchTimeout)
	}
}

func (tm *timeManager) ElapsedMilliseconds() int64 {
	return int64(time.Since(tm.start) / time.Millisecond)
}

func (tm *timeManager) IsSoftTimeout() bool {
	return tm.softTime > 0 && time.Since(tm.start) >= tm.softTime
}

func (tm *timeManager) Close() {
	if t := tm.timer; t != nil {
		t.Stop()
	}
}

func NewTimeManager(limits LimitsType, timeControlStrategy timeControlStrategy,
	side bool, ct *CancellationToken) *timeManager {
	var start = time.Now()

	if ct == nil {
		ct = &CancellationToken{}
	}

	var main, increment int
	if side {
		main, increment = limits.WhiteTime, limits.WhiteIncrement
	} else {
		main, increment = limits.BlackTime, limits.BlackIncrement
	}

	var softTime, hardTime int
	if limits.MoveTime > 0 {
		hardTime = limits.MoveTime
	} else if main > 0 {
		softTime, hardTime = timeControlStrategy(main, increment, limits.MovesToGo)
	}

	var timer *time.Timer
	if hardTime > 0 {
		timer = time.AfterFunc(time.Duration(hardTime)*time.Millisecond, func() {
			ct.Cancel()
		})
	}
	return &timeManager{
		start:    start,
		timer:    timer,
		ct:       ct,
		softTime: time.Duration(softTime) * time.Millisecond,
	}
}

func timeControlSmart(main, inc, moves int) (softLimit, hardLimit int) {
	const (
		MovesToGo       = 50
		LastMoveReserve = 300
	)

	if moves == 0 || moves > MovesToGo {
		moves = MovesToGo
	}

	var maxLimit = main - LastMoveReserve
	if moves > 1 {
		maxLimit = Min(maxLimit, main/2+inc)
	}

	var safeMoves = float64(moves) * (2 - float64(moves)/MovesToGo)
	softLimit = int(float64(main)/safeMoves) + inc
	hardLimit = softLimit * 4

	softLimit = Max(1, Min(maxLimit, softLimit))
	hardLimit = Max(1, Min(maxLimit, hardLimit))

	return
}
