package engine

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var searchTimeout = errors.New("search timeout")

type timeControlStrategy func(main, inc, moves int) (softLimit, hardLimit int)

type timeManager struct {
	start                       time.Time
	softTime                    time.Duration
	nodes, softNodes, hardNodes int64
	done                        <-chan struct{}
	cancel                      context.CancelFunc
}

func (tm *timeManager) Nodes() int64 {
	return tm.nodes
}

func (tm *timeManager) IsHardTimeout() bool {
	select {
	case <-tm.done:
		return true
	default:
	}
	if tm.hardNodes > 0 && tm.nodes >= tm.hardNodes {
		return true
	}
	return false
}

func (tm *timeManager) IncNodes() {
	var nodes = atomic.AddInt64(&tm.nodes, 1)
	if (nodes&63) == 63 && tm.IsHardTimeout() {
		panic(searchTimeout)
	}
}

func (tm *timeManager) ElapsedMilliseconds() int64 {
	return int64(time.Since(tm.start) / time.Millisecond)
}

func (tm *timeManager) IsSoftTimeout() bool {
	return (tm.softTime > 0 && time.Since(tm.start) >= tm.softTime) ||
		(tm.softNodes > 0 && tm.nodes >= tm.softNodes)
}

func (tm *timeManager) Close() {
	if tm.cancel != nil {
		tm.cancel()
	}
}

func NewTimeManager(limits LimitsType, timeControlStrategy timeControlStrategy,
	side bool, ctx context.Context) *timeManager {
	var start = time.Now()

	if timeControlStrategy == nil {
		timeControlStrategy = TimeControlBasic
	}

	if ctx == nil {
		ctx = context.Background()
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
		ctx, cancel = context.WithTimeout(ctx, time.Duration(hardTime)*time.Millisecond)
	}
	return &timeManager{
		start:     start,
		hardNodes: int64(hardNodes),
		softNodes: int64(softNodes),
		softTime:  time.Duration(softTime) * time.Millisecond,
		done:      ctx.Done(),
		cancel:    cancel,
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
