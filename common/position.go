package common

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	s "strings"
	"unicode"
)

type coloredPiece struct {
	Type int
	Side bool
}

var castleMask [64]int

func createPosition(board [64]coloredPiece, wtm bool,
	castleRights, ep, fifty int) (Position, bool) {
	var p = Position{
		WhiteMove:    wtm,
		CastleRights: castleRights,
		EpSquare:     ep,
		Rule50:       fifty,
		LastMove:     MoveEmpty,
	}

	for sq, piece := range board {
		if piece.Type != Empty {
			xorPiece(&p, piece.Type, piece.Side, sq)
		}
	}

	p.Key = p.computeKey()
	p.Checkers = p.computeCheckers()

	if !p.isLegal() {
		return Position{}, false
	}
	return p, true
}

func NewPositionFromFEN(fen string) (Position, error) {
	var tokens = s.Split(fen, " ")
	if len(tokens) <= 3 {
		return Position{}, fmt.Errorf("parse fen failed %v", fen)
	}

	var board [64]coloredPiece

	var i = 0
	for _, ch := range tokens[0] {
		if unicode.IsDigit(ch) {
			var n, _ = strconv.Atoi(string(ch))
			i += n
		} else if unicode.IsLetter(ch) {
			var pt = parsePiece(ch)
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

	var pos, isLegal = createPosition(board, whiteMove, cr, epSquare, rule50)
	if !isLegal {
		return Position{}, fmt.Errorf("parse fen failed %v", fen)
	}
	return pos, nil
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

			var pieceSide = (p.White & SquareMask[sq]) != 0
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
	var bb = SquareMask[sq]
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
	var bb = SquareMask[sq]
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
	result.Rule50 = src.Rule50 + 1
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

func (p *Position) PiecesByColor(side bool) uint64 {
	if side {
		return p.White
	}
	return p.Black
}

func xorPiece(p *Position, piece int, side bool, square int) {
	var b = SquareMask[square]
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
	var b = SquareMask[from] ^ SquareMask[to]
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
	var enemy = p.PiecesByColor(side)
	if (PawnAttacks(sq, !side) & p.Pawns & enemy) != 0 {
		return true
	}
	if (KnightAttacks[sq] & p.Knights & enemy) != 0 {
		return true
	}
	if (KingAttacks[sq] & p.Kings & enemy) != 0 {
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
		(KnightAttacks[sq] & p.Knights) |
		(BishopAttacks(sq, occ) & (p.Bishops | p.Queens)) |
		(RookAttacks(sq, occ) & (p.Rooks | p.Queens)) |
		(KingAttacks[sq] & p.Kings)
}

func (p *Position) computeCheckers() uint64 {
	if p.WhiteMove {
		return p.attackersTo(FirstOne(p.Kings&p.White)) & p.Black
	}
	return p.attackersTo(FirstOne(p.Kings&p.Black)) & p.White
}

func (p *Position) isLegal() bool {
	var kingSq = FirstOne(p.Kings & p.PiecesByColor(!p.WhiteMove))
	return !p.isAttackedBySide(kingSq, p.WhiteMove)
}

func (p *Position) IsCheck() bool {
	return p.Checkers != 0
}

func (p *Position) IsDiscoveredCheck() bool {
	return (p.Checkers & ^SquareMask[p.LastMove.To()]) != 0
}

func (p *Position) IsRepetition(other *Position) bool {
	return p.White == other.White &&
		p.Black == other.Black &&
		p.Pawns == other.Pawns &&
		p.Knights == other.Knights &&
		p.Bishops == other.Bishops &&
		p.Rooks == other.Rooks &&
		p.Queens == other.Queens &&
		p.Kings == other.Kings &&
		p.WhiteMove == other.WhiteMove &&
		p.CastleRights == other.CastleRights &&
		p.EpSquare == other.EpSquare
}

var (
	sideKey        uint64
	enpassantKey   [8]uint64
	castlingKey    [16]uint64
	pieceSquareKey [7 * 2 * 64]uint64
)

func PieceSquareKey(piece int, side bool, square int) uint64 {
	return pieceSquareKey[MakePiece(piece, side)*64+square]
}

func (p *Position) computeKey() uint64 {
	var result = uint64(0)
	if p.WhiteMove {
		result ^= sideKey
	}
	result ^= castlingKey[p.CastleRights]
	if p.EpSquare != SquareNone {
		result ^= enpassantKey[File(p.EpSquare)]
	}
	for i := 0; i < 64; i++ {
		var piece = p.WhatPiece(i)
		if piece != Empty {
			var side = (p.White & SquareMask[i]) != 0
			result ^= PieceSquareKey(piece, side, i)
		}
	}
	return result
}

func initKeys() {
	var r = rand.New(rand.NewSource(0))
	sideKey = r.Uint64()
	for i := range enpassantKey {
		enpassantKey[i] = r.Uint64()
	}
	for i := range pieceSquareKey {
		pieceSquareKey[i] = r.Uint64()
	}

	var castle [4]uint64
	for i := range castle {
		castle[i] = r.Uint64()
	}

	for i := range castlingKey {
		for j := 0; j < 4; j++ {
			if (i & (1 << uint(j))) != 0 {
				castlingKey[i] ^= castle[j]
			}
		}
	}
}

func MirrorPosition(p *Position) Position {
	var board [64]coloredPiece
	for i := range board {
		var pt, side = p.GetPieceTypeAndSide(i)
		if pt != Empty {
			board[FlipSquare(i)] = coloredPiece{pt, !side}
		}
	}
	var cr = (p.CastleRights >> 2) | ((p.CastleRights & 3) << 2)
	var ep = SquareNone
	if p.EpSquare != SquareNone {
		ep = FlipSquare(p.EpSquare)
	}
	var pos, _ = createPosition(board, !p.WhiteMove, cr, ep, p.Rule50)
	return pos
}

func init() {
	initKeys()
	for i := range castleMask {
		castleMask[i] = WhiteKingSide | WhiteQueenSide | BlackKingSide | BlackQueenSide
	}
	castleMask[SquareA1] &^= WhiteQueenSide
	castleMask[SquareE1] &^= WhiteQueenSide | WhiteKingSide
	castleMask[SquareH1] &^= WhiteKingSide
	castleMask[SquareA8] &^= BlackQueenSide
	castleMask[SquareE8] &^= BlackQueenSide | BlackKingSide
	castleMask[SquareH8] &^= BlackKingSide
}
