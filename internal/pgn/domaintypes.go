package pgn

import "github.com/ChizhovVadim/CounterGo/pkg/common"

const (
	GameResultNone     = "*"
	GameResultWhiteWin = "1-0"
	GameResultBlackWin = "0-1"
	GameResultDraw     = "1/2-1/2"
)

type Tag struct {
	Key   string
	Value string
}

type GameRaw struct {
	Tags    []Tag
	BodyRaw string
}

type Game struct {
	Result string
	Fen    string
	Items  []Item
}

type Comment struct {
	Depth int
	Score common.UciScore
}

type Item struct {
	Move common.Move
	Comment
}
