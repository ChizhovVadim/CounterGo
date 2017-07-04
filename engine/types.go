package engine

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
	MAX_HEIGHT                = 64
	MAX_MOVES                 = 256
	VALUE_DRAW                = 0
	VALUE_MATE                = 30000
	VALUE_INFINITE            = 30001
	VALUE_MATE_IN_MAX_HEIGHT  = VALUE_MATE - MAX_HEIGHT
	VALUE_MATED_IN_MAX_HEIGHT = -VALUE_MATE + MAX_HEIGHT
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

	A1 = 0
	B1 = 1
	C1 = 2
	D1 = 3
	E1 = 4
	F1 = 5
	G1 = 6
	H1 = 7

	A8 = 56
	B8 = 57
	C8 = 58
	D8 = 59
	E8 = 60
	F8 = 61
	G8 = 62
	H8 = 63
)

type Move int32

const MoveEmpty Move = 0

type MoveList struct {
	Items [MAX_MOVES]struct {
		Move  Move
		Score int
	}
	Count int
}

type SearchStack struct {
	Previous           *SearchStack
	Next               *SearchStack
	Position           *Position
	MoveList           *MoveList
	KillerMove         Move
	PrincipalVariation []Move
	QuietsSearched     []Move
}

type LimitsType struct {
	Ponder         bool
	Infinite       bool
	IsNodeLimits   bool
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
	IsTraceEnabled    bool
	Progress          func(si SearchInfo)
}

type SearchInfo struct {
	Score    int
	Depth    int
	Nodes    int64
	Time     int64
	MainLine []Move
}

type EvaluationFunc func(p *Position) int
