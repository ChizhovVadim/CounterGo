package engine

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	. "github.com/ChizhovVadim/CounterGo/common"
)

var searchTimeout = errors.New("search timeout")

type timeControlStrategy func(main, inc, moves int) (softLimit, hardLimit int)

type timeManager struct {
	start                   time.Time
	softTime                time.Duration
	nodes                   int64
	isCancellationRequested bool
}

func (tm *timeManager) Nodes() int64 {
	return tm.nodes
}

func (tm *timeManager) IsHardTimeout() bool {
	return tm.isCancellationRequested
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

func NewTimeManager(ctx context.Context, limits LimitsType,
	timeControlStrategy timeControlStrategy, side bool) (*timeManager, context.CancelFunc) {
	var start = time.Now()

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
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	var tm = &timeManager{
		start:    start,
		softTime: time.Duration(softTime) * time.Millisecond,
	}

	go func() {
		<-ctx.Done()
		tm.isCancellationRequested = true
	}()

	return tm, cancel
}

func timeControlSmart(main, inc, moves int) (softLimit, hardLimit int) {
	const (
		MovesToGo       = 35
		LastMoveReserve = 300
	)

	if moves == 0 || moves > MovesToGo {
		moves = MovesToGo
	}

	var maxLimit = main - LastMoveReserve
	if moves > 1 {
		maxLimit = Min(maxLimit, main/2+inc)
	}

	var safeMoves = 1 + float64(moves-1)*1.41
	softLimit = int(float64(main)/safeMoves) + inc
	hardLimit = softLimit * 4

	softLimit = Max(1, Min(maxLimit, softLimit))
	hardLimit = Max(1, Min(maxLimit, hardLimit))

	return
}
