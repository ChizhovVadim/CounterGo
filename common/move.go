package common

import "strings"

type Move int32

const MoveEmpty = Move(0)

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
