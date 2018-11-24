package engine

import (
	"math"
	"time"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	bestMoveChangedTimeRatio = 1.5
	bestMoveChangedStepCount = 3
)

var (
	bestMoveChangedK = math.Log(bestMoveChangedTimeRatio)
	timeRatioMin     = math.Exp(-0.5 * bestMoveChangedK)
	timeRatioMax     = math.Exp(0.5 * bestMoveChangedK)
	timeRatioMult    = math.Exp(-bestMoveChangedK / bestMoveChangedStepCount)
)

type timeControlStrategy func(main, inc, moves int) (softLimit, hardLimit int)

type timeManager struct {
	start    time.Time
	softTime time.Duration
	hardTime time.Duration
	ratio    float64
}

func (tm *timeManager) ElapsedMilliseconds() int64 {
	return int64(time.Since(tm.start) / time.Millisecond)
}

func (tm *timeManager) IsSoftTimeout(bestMoveChanged, problem bool) bool {
	if bestMoveChanged {
		tm.ratio = timeRatioMax
	} else {
		tm.ratio *= timeRatioMult
		if tm.ratio < timeRatioMin {
			tm.ratio = timeRatioMin
		}
	}
	if tm.softTime == 0 {
		return false
	}
	if problem {
		return false
	}
	return float64(time.Since(tm.start)) >= tm.ratio*float64(tm.softTime)
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
		ratio:    1.0,
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

	var safeMoves = 1 + 1.41*float64(moves-1)
	softLimit = int(float64(main)/safeMoves) + inc
	hardLimit = 4 * softLimit

	softLimit = Max(1, Min(maxLimit, softLimit))
	hardLimit = Max(1, Min(maxLimit, hardLimit))

	return
}
