package engine

import (
	"errors"
	"sync/atomic"
	"time"
)

type CancellationToken struct {
	active bool
}

func (ct *CancellationToken) Cancel() {
	ct.active = true
}

func (ct *CancellationToken) IsCancellationRequested() bool {
	return ct.active
}

var searchTimeout = errors.New("search timeout")

type TimeManagement struct {
	start                       time.Time
	softTime                    time.Duration
	nodes, softNodes, hardNodes int64
	ct                          *CancellationToken
	timer                       *time.Timer
}

func (tm *TimeManagement) Nodes() int64 {
	return tm.nodes
}

func (tm *TimeManagement) IncNodes() {
	atomic.AddInt64(&tm.nodes, 1)
}

func (tm *TimeManagement) ElapsedMilliseconds() int64 {
	return int64(time.Since(tm.start) / time.Millisecond)
}

func (tm *TimeManagement) PanicOnHardTimeout() {
	if tm.ct.IsCancellationRequested() ||
		(tm.hardNodes > 0 && tm.nodes >= tm.hardNodes) {
		panic(searchTimeout)
	}
}

func (tm *TimeManagement) IsSoftTimeout() bool {
	return (tm.softTime > 0 && time.Since(tm.start) >= tm.softTime) ||
		(tm.softNodes > 0 && tm.nodes >= tm.softNodes)
}

func (tm *TimeManagement) Close() {
	if t := tm.timer; t != nil {
		t.Stop()
	}
}

func NewTimeManagement(limits LimitsType, side bool, ct *CancellationToken) *TimeManagement {
	var start = time.Now()
	if ct == nil {
		ct = &CancellationToken{}
	}
	var hardNodes, softNodes int
	if limits.Nodes > 0 {
		hardNodes = limits.Nodes
	} else {
		var nodesPerGame = let(side, limits.WhiteNodesPerGame, limits.BlackNodesPerGame)
		if nodesPerGame > 0 {
			softNodes = nodesPerGame / 50
			hardNodes = min(nodesPerGame/2, 5*softNodes)
		}
	}
	var softTime, hardTime = ComputeThinkTime(limits, side)
	var timer *time.Timer
	if hardTime > 0 {
		timer = time.AfterFunc(time.Duration(hardTime)*time.Millisecond, func() {
			ct.Cancel()
		})
	}
	return &TimeManagement{
		start:     start,
		timer:     timer,
		ct:        ct,
		hardNodes: int64(hardNodes),
		softNodes: int64(softNodes),
		softTime:  time.Duration(softTime) * time.Millisecond,
	}
}

func ComputeThinkTime(limits LimitsType, side bool) (softLimit, hardLimit int) {
	const (
		MovesToGoDefault = 50
		MoveOverhead     = 20
	)
	if limits.MoveTime != 0 {
		return limits.MoveTime, limits.MoveTime
	}
	if limits.Infinite || limits.Ponder {
		return 0, 0
	}
	var mainTime, incTime int
	if side {
		mainTime, incTime = limits.WhiteTime, limits.WhiteIncrement
	} else {
		mainTime, incTime = limits.BlackTime, limits.BlackIncrement
	}
	var movesToGo int
	if 0 < limits.MovesToGo && limits.MovesToGo < MovesToGoDefault {
		movesToGo = limits.MovesToGo
	} else {
		movesToGo = MovesToGoDefault
	}

	var reserve = max(2*MoveOverhead, min(1000, mainTime/20))
	mainTime = max(0, mainTime-reserve)

	softLimit = mainTime/movesToGo + incTime
	hardLimit = min(mainTime/2, softLimit*5)
	return
}
