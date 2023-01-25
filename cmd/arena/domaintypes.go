package main

import (
	"context"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	gameResultDraw = iota
	gameResultWhiteWins
	gameResultBlackWins
)

type IEngine interface {
	Clear()
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

type timeControl struct {
	FixedNodes int
	FixedTime  time.Duration
}

type gameInfo struct {
	opening        string
	engineAIsWhite bool
	gameNumber     int
}

type gameResult struct {
	gameInfo  gameInfo
	positions []common.Position
	comment   string
	result    int
}
