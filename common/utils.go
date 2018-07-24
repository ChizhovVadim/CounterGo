package common

import (
	"strings"
	"unicode"
)

func Min(l, r int) int {
	if l < r {
		return l
	}
	return r
}

func Max(l, r int) int {
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
	return Max(FileDistance(sq1, sq2), RankDistance(sq1, sq2))
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

func parsePiece(ch rune) coloredPiece {
	var side = unicode.IsUpper(ch)
	var spiece = string(unicode.ToLower(ch))
	var i = strings.Index("pnbrqk", spiece)
	if i < 0 {
		return coloredPiece{Empty, false}
	}
	return coloredPiece{i + Pawn, side}
}

func makeMove(from, to, movingPiece, capturedPiece int) Move {
	return Move(from ^ (to << 6) ^ (movingPiece << 12) ^ (capturedPiece << 15))
}

func makePawnMove(from, to, capturedPiece, promotion int) Move {
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

func (p *Position) MakeMoveLAN(lan string) (Position, bool) {
	var buffer [MaxMoves]OrderedMove
	var ml = p.GenerateMoves(buffer[:])
	for i := range ml {
		var mv = ml[i].Move
		if strings.EqualFold(mv.String(), lan) {
			var newPosition = Position{}
			if p.MakeMove(mv, &newPosition) {
				return newPosition, true
			} else {
				return Position{}, false
			}
		}
	}
	return Position{}, false
}

func moveToSAN(pos *Position, ml []Move, mv Move) string {
	const PieceNames = "NBRQK"
	if mv == whiteKingSideCastle || mv == blackKingSideCastle {
		return "O-O"
	}
	if mv == whiteQueenSideCastle || mv == blackQueenSideCastle {
		return "O-O-O"
	}
	var strPiece, strCapture, strFrom, strTo, strPromotion string
	if mv.MovingPiece() != Pawn {
		strPiece = string(PieceNames[mv.MovingPiece()-Knight])
	}
	strTo = SquareName(mv.To())
	if mv.CapturedPiece() != Empty {
		strCapture = "x"
		if mv.MovingPiece() == Pawn {
			strFrom = SquareName(mv.From())[:1]
		}
	}
	if mv.Promotion() != Empty {
		strPromotion = "=" + string(PieceNames[mv.Promotion()-Knight])
	}
	var ambiguity = false
	var uniqCol = true
	var uniqRow = true
	for _, mv1 := range ml {
		if mv1.From() == mv.From() {
			continue
		}
		if mv1.To() != mv.To() {
			continue
		}
		if mv1.MovingPiece() != mv.MovingPiece() {
			continue
		}
		ambiguity = true
		if File(mv1.From()) == File(mv.From()) {
			uniqCol = false
		}
		if Rank(mv1.From()) == Rank(mv.From()) {
			uniqRow = false
		}
	}
	if ambiguity {
		if uniqCol {
			strFrom = SquareName(mv.From())[:1]
		} else if uniqRow {
			strFrom = SquareName(mv.From())[1:2]
		} else {
			strFrom = SquareName(mv.From())
		}
	}
	return strPiece + strFrom + strCapture + strTo + strPromotion
}

func ParseMoveSAN(pos *Position, san string) Move {
	var index = strings.IndexAny(san, "+#?!")
	if index >= 0 {
		san = san[:index]
	}
	var ml = pos.GenerateLegalMoves()
	for _, mv := range ml {
		if san == moveToSAN(pos, ml, mv) {
			return mv
		}
	}
	return MoveEmpty
}
