package engine

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var searchTimeout = errors.New("search timeout")

type TimeControlStrategy func(main, inc, moves int) (softLimit, hardLimit int)

type TimeManagement struct {
	start                       time.Time
	softTime                    time.Duration
	nodes, softNodes, hardNodes int64
	cancel                      context.CancelFunc
	ct                          context.Context
}

func (tm *TimeManagement) Nodes() int64 {
	return tm.nodes
}

func (tm *TimeManagement) IncNodes() {
	atomic.AddInt64(&tm.nodes, 1)
	if tm.hardNodes > 0 && tm.nodes >= tm.hardNodes {
		tm.cancel()
	}
}

func (tm *TimeManagement) ElapsedMilliseconds() int64 {
	return int64(time.Since(tm.start) / time.Millisecond)
}

func (tm *TimeManagement) IsSoftTimeout() bool {
	return (tm.softTime > 0 && time.Since(tm.start) >= tm.softTime) ||
		(tm.softNodes > 0 && tm.nodes >= tm.softNodes)
}

func (tm *TimeManagement) PanicOnHardTimeout() {
	select {
	case <-tm.ct.Done():
		panic(searchTimeout)
	default:
	}
}

func NewTimeManagement(limits LimitsType, timeControlStrategy TimeControlStrategy,
	side bool, ct context.Context) *TimeManagement {
	var start = time.Now()

	if timeControlStrategy == nil {
		timeControlStrategy = TimeControlBasic
	}

	if ct == nil {
		ct = context.Background()
	}

	var main, increment int
	if side {
		main, increment = limits.WhiteTime, limits.WhiteIncrement
	} else {
		main, increment = limits.BlackTime, limits.BlackIncrement
	}

	var softTime, hardTime, softNodes, hardNodes int
	if limits.MoveTime > 0 {
		hardTime = limits.MoveTime
	} else if limits.Nodes > 0 {
		hardNodes = limits.Nodes
	} else if main > 0 {
		var softLimit, hardLimit = timeControlStrategy(main, increment, limits.MovesToGo)
		if limits.IsNodeLimits {
			softNodes, hardNodes = softLimit, hardLimit
		} else {
			softTime, hardTime = softLimit, hardLimit
		}
	}

	var cancel context.CancelFunc
	if hardTime > 0 {
		ct, cancel = context.WithTimeout(ct, time.Duration(hardTime)*time.Millisecond)
	} else if hardNodes > 0 {
		ct, cancel = context.WithCancel(ct)
	}

	return &TimeManagement{
		start:     start,
		hardNodes: int64(hardNodes),
		softNodes: int64(softNodes),
		softTime:  time.Duration(softTime) * time.Millisecond,
		cancel:    cancel,
		ct:        ct,
	}
}

func computeLimit(main, inc, moves int) int {
	return (main + inc*(moves-1)) / moves
}

func TimeControlBasic(main, inc, moves int) (softLimit, hardLimit int) {
	const (
		SoftMovesToGo   = 50
		HardMovesToGo   = 10
		LastMoveReserve = 300
		MoveReserve     = 20
	)

	if moves == 0 {
		moves = SoftMovesToGo
	}

	softLimit = computeLimit(main, inc, min(moves, SoftMovesToGo))
	hardLimit = computeLimit(main, inc, min(moves, HardMovesToGo))

	hardLimit -= MoveReserve
	hardLimit = min(hardLimit, main-LastMoveReserve)
	hardLimit = max(hardLimit, 1)

	return
}
