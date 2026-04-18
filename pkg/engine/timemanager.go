package engine

import (
	"math"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	maxDifficulty   = 2
	minBranchFactor = 0.75
	maxBranchFactor = 1.5
)

type timeManager struct {
	start        time.Time
	limits       common.LimitsType
	side         bool
	difficulty   float64
	lastScore    int
	lastBestMove common.Move
}

func newTimeManager(start time.Time, limits common.LimitsType, p *common.Position) *timeManager {
	return &timeManager{
		start:      start,
		limits:     limits,
		side:       p.WhiteMove,
		difficulty: 1,
	}
}

func (tm *timeManager) HardLimit() time.Duration {
	if tm.limits.MoveTime > 0 {
		return time.Duration(tm.limits.MoveTime) * time.Millisecond
	} else if tm.limits.WhiteTime > 0 || tm.limits.BlackTime > 0 {
		return tm.calculateTimeLimit(maxDifficulty, maxBranchFactor)
	} else {
		return 0
	}
}

func (tm *timeManager) IsFixedNodesInterruption(nodes int64) bool {
	return tm.limits.Nodes > 0 && nodes >= int64(tm.limits.Nodes)
}

func (tm *timeManager) IsSoftInterruption(line mainLine) bool {
	if tm.limits.Infinite {
		return false
	}
	if tm.limits.Depth != 0 && line.depth >= tm.limits.Depth {
		return true
	}
	if line.score >= winIn(line.depth-5) ||
		line.score <= lossIn(line.depth-5) {
		return true
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
			return true
		}
	}
	return false
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
