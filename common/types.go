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

type OrderedMove struct {
	Move Move
	Key  int
}

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
	Positions []Position
	Limits    LimitsType
	Progress  func(si SearchInfo)
}

type SearchInfo struct {
	Score    UciScore
	Depth    int
	Nodes    int64
	Time     int64
	MainLine []Move
}

type UciScore struct {
	Centipawns int
	Mate       int
}
