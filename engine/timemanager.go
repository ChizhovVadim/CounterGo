package engine

import (
	"time"

	. "github.com/ChizhovVadim/CounterGo/common"
)

type timeControlStrategy func(main, inc, moves int) (softLimit, hardLimit int)

type timeManager struct {
	start    time.Time
	softTime time.Duration
	hardTime time.Duration
}

func (tm *timeManager) ElapsedMilliseconds() int64 {
	return int64(time.Since(tm.start) / time.Millisecond)
}

func (tm *timeManager) IsSoftTimeout() bool {
	return tm.softTime > 0 && time.Since(tm.start) >= tm.softTime
}

func NewTimeManager(limits LimitsType,
	timeControlStrategy timeControlStrategy, side bool) timeManager {
	var start = time.Now()

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

	return timeManager{
		start:    start,
		softTime: time.Duration(softTime) * time.Millisecond,
		hardTime: time.Duration(hardTime) * time.Millisecond,
	}
}

func timeControlSmart(main, inc, moves int) (softLimit, hardLimit int) {
	const (
		MovesToGo       = 35
		LastMoveReserve = 300
	)

	if moves == 0 || moves > MovesToGo {
		moves = MovesToGo
	}

	main = Max(1, main-LastMoveReserve)
	var maxLimit = main
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
