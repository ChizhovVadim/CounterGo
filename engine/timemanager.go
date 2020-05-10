package engine

import (
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
}

func (tm *timeManager) Init(start time.Time, limits LimitsType, p *Position) {
	tm.start = start
	tm.limits = limits
	tm.side = p.WhiteMove
	tm.difficulty = 1
}

func (tm *timeManager) Deadline() (time.Time, bool) {
	if tm.limits.Infinite {
		return time.Time{}, false
	}
	var maximum time.Duration
	if tm.limits.MoveTime > 0 {
		maximum = time.Duration(tm.limits.MoveTime) * time.Millisecond
	} else {
		maximum = tm.calculateTimeLimit(maxDifficulty, maxBranchFactor)
	}
	return tm.start.Add(maximum), true
}

func (tm *timeManager) BreakIterativeDeepening(line mainLine) bool {
	if tm.limits.MoveTime > 0 || tm.limits.Infinite {
		return false
	}
	if line.score >= winIn(line.depth-3) ||
		line.score <= lossIn(line.depth-3) {
		return true
	}
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
	return time.Since(tm.start) >= optimum
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
