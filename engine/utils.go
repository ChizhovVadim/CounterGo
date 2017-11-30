package engine

import (
	"strings"
	"unicode"
)

func initSliceInt(size int, f func(index int) int) []int {
	var result = make([]int, size)
	for i := range result {
		result[i] = f(i)
	}
	return result
}

func interpolateLinearInt(x, x_min, x_max, y_min, y_max int) int {
	return ((y_max-y_min)*(x-x_min)+(x_max-x_min)/2)/(x_max-x_min) + y_min
}

func min(l, r int) int {
	if l < r {
		return l
	}
	return r
}

func max(l, r int) int {
	if l > r {
		return l
	}
	return r
}

func let(ok bool, yes, no int) int {
	if ok {
		return yes
	}
	return no
}

func FlipSquare(sq int) int {
	return sq ^ 56
}

func File(sq int) int {
	return sq & 7
}

func Rank(sq int) int {
	return sq >> 3
}

func IsDarkSquare(sq int) bool {
	return (File(sq) & 1) == (Rank(sq) & 1)
}

func AbsDelta(x, y int) int {
	if x > y {
		return x - y
	}
	return y - x
}

func FileDistance(sq1, sq2 int) int {
	return AbsDelta(File(sq1), File(sq2))
}

func RankDistance(sq1, sq2 int) int {
	return AbsDelta(Rank(sq1), Rank(sq2))
}

func SquareDistance(sq1, sq2 int) int {
	return max(FileDistance(sq1, sq2), RankDistance(sq1, sq2))
}

func MakeSquare(file, rank int) int {
	return (rank << 3) | file
}

const (
	fileNames = "abcdefgh"
	rankNames = "12345678"
)

func SquareName(sq int) string {
	var file = fileNames[File(sq)]
	var rank = rankNames[Rank(sq)]
	return string(file) + string(rank)
}

func ParseSquare(s string) int {
	if s == "-" {
		return SquareNone
	}
	var file = strings.Index(fileNames, s[0:1])
	var rank = strings.Index(rankNames, s[1:2])
	return MakeSquare(file, rank)
}

func ParsePiece(ch rune) int {
	var side = unicode.IsUpper(ch)
	var spiece = string(unicode.ToLower(ch))
	var i = strings.Index("pnbrqk", spiece)
	if i < 0 {
		return Empty
	}
	return MakePiece(i+Pawn, side)
}

func MakeMove(from, to, movingPiece, capturedPiece int) Move {
	return Move(from ^ (to << 6) ^ (movingPiece << 12) ^ (capturedPiece << 15))
}

func MakePawnMove(from, to, capturedPiece, promotion int) Move {
	return Move(from ^ (to << 6) ^ (Pawn << 12) ^ (capturedPiece << 15) ^ (promotion << 18))
}

func (m Move) From() int {
	return int(m & 63)
}

func (m Move) To() int {
	return int((m >> 6) & 63)
}

func (m Move) MovingPiece() int {
	return int((m >> 12) & 7)
}

func (m Move) CapturedPiece() int {
	return int((m >> 15) & 7)
}

func (m Move) Promotion() int {
	return int((m >> 18) & 7)
}

func MakePiece(pieceType int, side bool) int {
	if side {
		return pieceType
	}
	return pieceType + 7
}

func GetPieceTypeAndSide(piece int) (pieceType int, side bool) {
	if piece < 7 {
		return piece, true
	} else {
		return piece - 7, false
	}
}

func (m Move) String() string {
	if m == MoveEmpty {
		return "0000"
	}
	var sPromotion = ""
	if m.Promotion() != Empty {
		sPromotion = string("nbrq"[m.Promotion()-Knight])
	}
	return SquareName(m.From()) + SquareName(m.To()) + sPromotion
}

func ParseMove(s string) Move {
	s = strings.ToLower(s)
	var from = ParseSquare(s[0:2])
	var to = ParseSquare(s[2:4])
	if len(s) <= 4 {
		return MakeMove(from, to, Empty, Empty)
	}
	var promotion = strings.Index("nbrqk", strings.ToLower(s[4:5])) + Knight
	return MakePawnMove(from, to, Empty, promotion)
}
