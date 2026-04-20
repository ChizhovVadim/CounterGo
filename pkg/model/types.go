package model

import (
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type LimitsType struct {
	Ponder         bool
	Infinite       bool
	WhiteTime      int
	BlackTime      int
	WhiteIncrement int
	BlackIncrement int
	MoveTime       int
	MovesToGo      int
	Depth          int
	Nodes          int
	Mate           int
}

type SearchParams struct {
	Game     Game
	Limits   LimitsType
	Progress func(si SearchInfo)
}

type SearchInfo struct {
	Score    UciScore
	Depth    int
	Nodes    int64
	Time     time.Duration
	MainLine []common.Move
}

type UciScore struct {
	Centipawns int
	Mate       int
}
