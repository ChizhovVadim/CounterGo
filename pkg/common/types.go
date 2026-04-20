package common

const (
	SQUARE_NB = 64
	RANK_NB   = 8
	FILE_NB   = 8
	COLOUR_NB = 2
	PIECE_NB  = 8
)

const (
	SideWhite = 0
	SideBlack = 1
)

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
	Key  int32
}
