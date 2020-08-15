package engine

import (
	"context"
	"math"
	"time"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	maxDifficulty   = 2
	minBranchFactor = 0.75
	maxBranchFactor = 1.5
)

type timeManager struct {
	start        time.Time
	limits       LimitsType
	side         bool
	difficulty   float64
	lastScore    int
	lastBestMove Move
	cancel       context.CancelFunc
}

func withTimeManager(ctx context.Context, start time.Time,
	limits LimitsType, p *Position) (context.Context, *timeManager) {

	var tm = &timeManager{
		start:      start,
		limits:     limits,
		side:       p.WhiteMove,
		difficulty: 1,
	}

	var cancel context.CancelFunc
	if limits.MoveTime > 0 || limits.WhiteTime > 0 || limits.BlackTime > 0 {
		var maximum time.Duration
		if limits.MoveTime > 0 {
			maximum = time.Duration(limits.MoveTime) * time.Millisecond
		} else {
			maximum = tm.calculateTimeLimit(maxDifficulty, maxBranchFactor)
		}
		ctx, cancel = context.WithDeadline(ctx, start.Add(maximum))
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	tm.cancel = cancel
	return ctx, tm
}

func (tm *timeManager) OnNodesChanged(nodes int) {
	if tm.limits.Nodes > 0 && nodes >= tm.limits.Nodes {
		tm.cancel()
	}
}

func (tm *timeManager) OnIterationComplete(line mainLine) {
	if tm.limits.Infinite {
		return
	}
	if line.score >= winIn(line.depth-3) ||
		line.score <= lossIn(line.depth-3) {
		tm.cancel()
		return
	}
	if tm.limits.WhiteTime > 0 || tm.limits.BlackTime > 0 {
		if line.depth >= 5 {
			if line.score < tm.lastScore-pawnValue/2 {
				tm.difficulty = maxDifficulty
			} else if line.moves[0] != tm.lastBestMove {
				tm.difficulty = math.Max(1.5, tm.difficulty)
			} else {
				tm.difficulty = math.Max(0.95, 0.9*tm.difficulty)
			}
		}
		tm.lastScore = line.score
		tm.lastBestMove = line.moves[0]
		var optimum = tm.calculateTimeLimit(tm.difficulty, minBranchFactor)
		if time.Since(tm.start) >= optimum {
			tm.cancel()
			return
		}
	}
}

func (tm *timeManager) Close() {
	tm.cancel()
}

func (tm *timeManager) calculateTimeLimit(difficulty, branchFactor float64) time.Duration {
	const (
		DefaultMovesToGo = 40
		MoveOverhead     = 300 * time.Millisecond
		MinTimeLimit     = 1 * time.Millisecond
	)
	var main, inc time.Duration
	if tm.side {
		main = time.Duration(tm.limits.WhiteTime) * time.Millisecond
		inc = time.Duration(tm.limits.WhiteIncrement) * time.Millisecond
	} else {
		main = time.Duration(tm.limits.BlackTime) * time.Millisecond
		inc = time.Duration(tm.limits.BlackIncrement) * time.Millisecond
	}
	main -= MoveOverhead
	if main < MinTimeLimit {
		main = MinTimeLimit
	}
	var moves = tm.limits.MovesToGo
	if moves == 0 || moves > DefaultMovesToGo {
		moves = DefaultMovesToGo
	}
	var total = float64(main) + float64(moves-1)*float64(inc)
	var timeLimit = time.Duration(difficulty * branchFactor * total / (difficulty*maxBranchFactor + float64(moves-1)))
	if timeLimit > main {
		timeLimit = main
	}
	if timeLimit < MinTimeLimit {
		timeLimit = MinTimeLimit
	}
	return timeLimit
}
