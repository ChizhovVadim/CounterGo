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
	start    time.Time
	softTime time.Duration
	nodes    int64
	done     <-chan struct{}
	cancel   context.CancelFunc
}

func (tm *timeManager) Nodes() int64 {
	return tm.nodes
}

func (tm *timeManager) IsHardTimeout() bool {
	select {
	case <-tm.done:
		return true
	default:
		return false
	}
}

func (tm *timeManager) IncNodes() {
	var nodes = atomic.AddInt64(&tm.nodes, 1)
	if (nodes&63) == 0 && tm.IsHardTimeout() {
		panic(searchTimeout)
	}
}

func (tm *timeManager) ElapsedMilliseconds() int64 {
	return int64(time.Since(tm.start) / time.Millisecond)
}

func (tm *timeManager) IsSoftTimeout() bool {
	return (tm.softTime > 0 && time.Since(tm.start) >= tm.softTime)
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
		timeControlStrategy = timeControlSmart
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

	var softTime, hardTime int
	if limits.MoveTime > 0 {
		hardTime = limits.MoveTime
	} else if main > 0 {
		softTime, hardTime = timeControlStrategy(main, increment, limits.MovesToGo)
	}

	var cancel context.CancelFunc
	if hardTime > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(hardTime)*time.Millisecond)
	}
	return &timeManager{
		start:    start,
		softTime: time.Duration(softTime) * time.Millisecond,
		done:     ctx.Done(),
		cancel:   cancel,
	}
}

func timeControlSmart(main, inc, moves int) (softLimit, hardLimit int) {
	const (
		MovesToGo       = 25
		LastMoveReserve = 300
		MoveReserve     = 20
	)

	if moves == 0 || moves > MovesToGo {
		moves = MovesToGo
	}

	var total = main + inc*(moves-1)
	var alloc = total / moves
	if moves > 1 {
		total /= 2
	}
	var maxLimit = min(main-LastMoveReserve, total)
	softLimit = max(1, min(alloc/2, maxLimit))
	hardLimit = max(1, min(alloc*2-MoveReserve, maxLimit))

	return
}
