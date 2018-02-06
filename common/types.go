package common

const (
	WhiteKingSide = 1 << iota
	WhiteQueenSide
	BlackKingSide
	BlackQueenSide
)

type Position struct {
	Pawns, Knights, Bishops, Rooks, Queens, Kings, White, Black, Checkers uint64
	WhiteMove                                                             bool
	CastleRights, Rule50, EpSquare                                        int
	Key                                                                   uint64
	LastMove                                                              Move
}

const InitialPositionFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

const (
	Empty int = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

const (
	MaxMoves = 256
)

const (
	FileA = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

const (
	Rank1 = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

const (
	SquareNone = -1
	SquareA1   = 0
	SquareB1   = 1
	SquareC1   = 2
	SquareD1   = 3
	SquareE1   = 4
	SquareF1   = 5
	SquareG1   = 6
	SquareH1   = 7
	SquareA8   = 56
	SquareB8   = 57
	SquareC8   = 58
	SquareD8   = 59
	SquareE8   = 60
	SquareF8   = 61
	SquareG8   = 62
	SquareH8   = 63
)

type Move int32

const MoveEmpty Move = 0

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
	Positions         []*Position
	Limits            LimitsType
	CancellationToken *CancellationToken
	Progress          func(si SearchInfo)
}

type SearchInfo struct {
	Score    Score
	Depth    int
	Nodes    int64
	Time     int64
	MainLine []Move
}

type Score struct {
	Centipawns int
	Mate       int
}

type CancellationToken struct {
	active bool
}

func (ct *CancellationToken) Cancel() {
	ct.active = true
}

func (ct *CancellationToken) IsCancellationRequested() bool {
	return ct.active
}

type UciOption interface {
	GetName() string
}

type BoolUciOption struct {
	Name  string
	Value bool
}

func (o *BoolUciOption) GetName() string {
	return o.Name
}

type IntUciOption struct {
	Name            string
	Value, Min, Max int
}

func (o *IntUciOption) GetName() string {
	return o.Name
}
