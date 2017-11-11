package engine

import (
	"bytes"
	"fmt"
	"strconv"
	s "strings"
	"unicode"
)

var castleMask [64]int

func createPosition(board [64]int, wtm bool, castleRights, ep, fifty int) *Position {
	var p = &Position{}

	for i := 0; i < 64; i++ {
		if board[i] != Empty {
			var piece, pieceSide = GetPieceTypeAndSide(board[i])
			xorPiece(p, piece, pieceSide, i)
		}
	}

	p.WhiteMove = wtm
	p.CastleRights = castleRights
	p.EpSquare = ep
	p.Rule50 = fifty
	p.Key = p.ComputeKey()
	p.Checkers = p.computeCheckers()
	p.LastMove = MoveEmpty

	if !p.isLegal() {
		return nil
	}
	return p
}

func NewPositionFromFEN(fen string) *Position {
	var tokens = s.Split(fen, " ")

	var board [64]int

	var i = 0
	for _, ch := range tokens[0] {
		if unicode.IsDigit(ch) {
			var n, _ = strconv.Atoi(string(ch))
			i += n
		} else if unicode.IsLetter(ch) {
			var pt = ParsePiece(ch)
			board[FlipSquare(i)] = pt
			i++
		}
	}

	var whiteMove = tokens[1] == "w"

	var sCastleRights = tokens[2]
	var cr = 0
	if s.Contains(sCastleRights, "K") {
		cr |= WhiteKingSide
	}
	if s.Contains(sCastleRights, "Q") {
		cr |= WhiteQueenSide
	}
	if s.Contains(sCastleRights, "k") {
		cr |= BlackKingSide
	}
	if s.Contains(sCastleRights, "q") {
		cr |= BlackQueenSide
	}

	var epSquare = ParseSquare(tokens[3])

	var rule50 = 0
	if len(tokens) > 4 {
		rule50, _ = strconv.Atoi(tokens[4])
	}

	return createPosition(board, whiteMove, cr, epSquare, rule50)
}

func (p *Position) String() string {
	var sb bytes.Buffer

	var emptyCount = 0

	for i := 0; i < 64; i++ {
		var sq = FlipSquare(i)
		var piece = p.WhatPiece(sq)
		if piece == Empty {
			emptyCount++
		} else {
			if emptyCount != 0 {
				sb.WriteString(strconv.Itoa(emptyCount))
				emptyCount = 0
			}

			var pieceSide bool
			if (p.White & squareMask[sq]) != 0 {
				pieceSide = true
			} else {
				pieceSide = false
			}

			sb.WriteString(pieceToChar(piece, pieceSide))
		}

		if File(sq) == FileH {
			if emptyCount != 0 {
				sb.WriteString(strconv.Itoa(emptyCount))
				emptyCount = 0
			}
			if Rank(sq) != Rank1 {
				sb.WriteString("/")
			}
		}
	}
	sb.WriteString(" ")

	if p.WhiteMove {
		sb.WriteString("w")
	} else {
		sb.WriteString("b")
	}
	sb.WriteString(" ")

	if p.CastleRights == 0 {
		sb.WriteString("-")
	} else {
		if (p.CastleRights & WhiteKingSide) != 0 {
			sb.WriteString("K")
		}
		if (p.CastleRights & WhiteQueenSide) != 0 {
			sb.WriteString("Q")
		}
		if (p.CastleRights & BlackKingSide) != 0 {
			sb.WriteString("k")
		}
		if (p.CastleRights & BlackQueenSide) != 0 {
			sb.WriteString("q")
		}
	}
	sb.WriteString(" ")

	if p.EpSquare == SquareNone {
		sb.WriteString("-")
	} else {
		sb.WriteString(SquareName(p.EpSquare))
	}
	sb.WriteString(" ")

	sb.WriteString(strconv.Itoa(p.Rule50))
	sb.WriteString(" ")

	sb.WriteString(strconv.Itoa(p.Rule50/2 + 1))

	return sb.String()
}

func pieceToChar(pieceType int, side bool) string {
	var result = string("pnbrqk"[pieceType-Pawn])
	if side {
		result = s.ToUpper(result)
	}
	return result
}

func (p *Position) GetPieceTypeAndSide(sq int) (pieceType int, side bool) {
	var bb = squareMask[sq]
	if (p.White & bb) != 0 {
		side = true
	} else if (p.Black & bb) != 0 {
		side = false
	} else {
		pieceType = Empty
		return
	}
	pieceType = p.WhatPiece(sq)
	return
}

func (p *Position) WhatPiece(sq int) int {
	var bb = squareMask[sq]
	if ((p.White | p.Black) & bb) == 0 {
		return Empty
	}
	if (p.Pawns & bb) != 0 {
		return Pawn
	}
	if (p.Knights & bb) != 0 {
		return Knight
	}
	if (p.Bishops & bb) != 0 {
		return Bishop
	}

	if (p.Rooks & bb) != 0 {
		return Rook
	}
	if (p.Queens & bb) != 0 {
		return Queen
	}
	if (p.Kings & bb) != 0 {
		return King
	}
	panic(fmt.Errorf("Wrong piece on %s", SquareName(sq)))
}

func (src *Position) MakeMove(move Move, result *Position) bool {
	var from = move.From()
	var to = move.To()
	var movingPiece = move.MovingPiece()
	var capturedPiece = move.CapturedPiece()

	result.Pawns = src.Pawns
	result.Knights = src.Knights
	result.Bishops = src.Bishops
	result.Rooks = src.Rooks
	result.Queens = src.Queens
	result.Kings = src.Kings
	result.White = src.White
	result.Black = src.Black

	result.WhiteMove = !src.WhiteMove
	result.Key = src.Key ^ sideKey

	result.CastleRights = src.CastleRights & castleMask[from] & castleMask[to]
	result.Key ^= castlingKey[result.CastleRights^src.CastleRights]

	if movingPiece == Pawn || capturedPiece != Empty {
		result.Rule50 = 0
	} else {
		result.Rule50 = src.Rule50 + 1
	}

	result.EpSquare = SquareNone
	if src.EpSquare != SquareNone {
		result.Key ^= enpassantKey[File(src.EpSquare)]
	}

	if capturedPiece != Empty {
		if capturedPiece == Pawn && to == src.EpSquare {
			xorPiece(result, Pawn, !src.WhiteMove, to+let(src.WhiteMove, -8, 8))
		} else {
			xorPiece(result, capturedPiece, !src.WhiteMove, to)
		}
	}

	movePiece(result, movingPiece, src.WhiteMove, from, to)

	if movingPiece == Pawn {
		if src.WhiteMove {
			if to == from+16 {
				result.EpSquare = from + 8
				result.Key ^= enpassantKey[File(from+8)]
			}
			if Rank(to) == Rank8 {
				xorPiece(result, Pawn, true, to)
				xorPiece(result, move.Promotion(), true, to)
			}
		} else {
			if to == from-16 {
				result.EpSquare = from - 8
				result.Key ^= enpassantKey[File(from-8)]
			}
			if Rank(to) == Rank1 {
				xorPiece(result, Pawn, false, to)
				xorPiece(result, move.Promotion(), false, to)
			}
		}
	} else if movingPiece == King {
		if src.WhiteMove {
			if from == SquareE1 && to == SquareG1 {
				movePiece(result, Rook, true, SquareH1, SquareF1)
			}
			if from == SquareE1 && to == SquareC1 {
				movePiece(result, Rook, true, SquareA1, SquareD1)
			}
		} else {
			if from == SquareE8 && to == SquareG8 {
				movePiece(result, Rook, false, SquareH8, SquareF8)
			}
			if from == SquareE8 && to == SquareC8 {
				movePiece(result, Rook, false, SquareA8, SquareD8)
			}
		}
	}

	if !result.isLegal() {
		return false
	}
	result.Checkers = result.computeCheckers()
	result.LastMove = move
	return true
}

func (src *Position) MakeNullMove(result *Position) {
	result.Pawns = src.Pawns
	result.Knights = src.Knights
	result.Bishops = src.Bishops
	result.Rooks = src.Rooks
	result.Queens = src.Queens
	result.Kings = src.Kings
	result.White = src.White
	result.Black = src.Black
	result.Rule50 = src.Rule50
	result.CastleRights = src.CastleRights

	result.WhiteMove = !src.WhiteMove
	result.Key = src.Key ^ sideKey

	result.EpSquare = SquareNone
	if src.EpSquare != SquareNone {
		result.Key ^= enpassantKey[File(src.EpSquare)]
	}

	result.Checkers = 0
	result.LastMove = MoveEmpty
}

func (p *Position) piecesByColor(side bool) uint64 {
	if side {
		return p.White
	}
	return p.Black
}

func xorPiece(p *Position, piece int, side bool, square int) {
	var b = squareMask[square]
	if side {
		p.White ^= b
	} else {
		p.Black ^= b
	}
	switch piece {
	case Pawn:
		p.Pawns ^= b
	case Knight:
		p.Knights ^= b
	case Bishop:
		p.Bishops ^= b
	case Rook:
		p.Rooks ^= b
	case Queen:
		p.Queens ^= b
	case King:
		p.Kings ^= b
	}
	p.Key ^= PieceSquareKey(piece, side, square)
}

func movePiece(p *Position, piece int, side bool, from int, to int) {
	var b = squareMask[from] ^ squareMask[to]
	if side {
		p.White ^= b
	} else {
		p.Black ^= b
	}
	switch piece {
	case Pawn:
		p.Pawns ^= b
	case Knight:
		p.Knights ^= b
	case Bishop:
		p.Bishops ^= b
	case Rook:
		p.Rooks ^= b
	case Queen:
		p.Queens ^= b
	case King:
		p.Kings ^= b
	}
	p.Key ^= PieceSquareKey(piece, side, from) ^ PieceSquareKey(piece, side, to)
}

func (p *Position) isAttackedBySide(sq int, side bool) bool {
	var enemy = p.piecesByColor(side)
	if (PawnAttacks(sq, !side) & p.Pawns & enemy) != 0 {
		return true
	}
	if (knightAttacks[sq] & p.Knights & enemy) != 0 {
		return true
	}
	if (kingAttacks[sq] & p.Kings & enemy) != 0 {
		return true
	}
	var allPieces = p.White | p.Black
	if (BishopAttacks(sq, allPieces) & (p.Bishops | p.Queens) & enemy) != 0 {
		return true
	}
	if (RookAttacks(sq, allPieces) & (p.Rooks | p.Queens) & enemy) != 0 {
		return true
	}
	return false
}

func (p *Position) attackersTo(sq int) uint64 {
	var occ = p.White | p.Black
	return (blackPawnAttacks[sq] & p.Pawns & p.White) |
		(whitePawnAttacks[sq] & p.Pawns & p.Black) |
		(knightAttacks[sq] & p.Knights) |
		(BishopAttacks(sq, occ) & (p.Bishops | p.Queens)) |
		(RookAttacks(sq, occ) & (p.Rooks | p.Queens)) |
		(kingAttacks[sq] & p.Kings)
}

func (p *Position) computeCheckers() uint64 {
	if p.WhiteMove {
		return p.attackersTo(FirstOne(p.Kings&p.White)) & p.Black
	}
	return p.attackersTo(FirstOne(p.Kings&p.Black)) & p.White
}

func (p *Position) isLegal() bool {
	var kingSq = FirstOne(p.Kings & p.piecesByColor(!p.WhiteMove))
	return !p.isAttackedBySide(kingSq, p.WhiteMove)
}

func (p *Position) IsCheck() bool {
	return p.Checkers != 0
}

func (p *Position) IsDiscoveredCheck() bool {
	return (p.Checkers & ^squareMask[p.LastMove.To()]) != 0
}

func (p *Position) MakeMoveIfLegal(move Move) *Position {
	var moveList MoveList
	moveList.GenerateMoves(p)
	for i := 0; i < moveList.Count; i++ {
		var x = moveList.Items[i].Move
		if move.From() == x.From() && move.To() == x.To() && move.Promotion() == x.Promotion() {
			var newPosition = &Position{}
			if p.MakeMove(x, newPosition) {
				return newPosition
			} else {
				return nil
			}
		}
	}
	return nil
}

func (this *Position) IsRepetition(other *Position) bool {
	return this.White == other.White &&
		this.Black == other.Black &&
		this.Pawns == other.Pawns &&
		this.Knights == other.Knights &&
		this.Bishops == other.Bishops &&
		this.Rooks == other.Rooks &&
		this.Queens == other.Queens &&
		this.Kings == other.Kings &&
		this.WhiteMove == other.WhiteMove &&
		this.CastleRights == other.CastleRights &&
		this.EpSquare == other.EpSquare
}

func init() {
	for i := 0; i < len(castleMask); i++ {
		castleMask[i] = WhiteKingSide | WhiteQueenSide | BlackKingSide | BlackQueenSide
	}
	castleMask[SquareA1] &^= WhiteQueenSide
	castleMask[SquareE1] &^= WhiteQueenSide | WhiteKingSide
	castleMask[SquareH1] &^= WhiteKingSide
	castleMask[SquareA8] &^= BlackQueenSide
	castleMask[SquareE8] &^= BlackQueenSide | BlackKingSide
	castleMask[SquareH8] &^= BlackKingSide
}
