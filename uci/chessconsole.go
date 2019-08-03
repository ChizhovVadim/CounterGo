package uci

import (
	"fmt"
	"strconv"

	"github.com/ChizhovVadim/CounterGo/common"
)

const (
	whiteKing   = "\u2654"
	whiteQueen  = "\u2655"
	whiteRook   = "\u2656"
	whiteBishop = "\u2657"
	whiteKnight = "\u2658"
	whitePawn   = "\u2659"
	blackKing   = "\u265A"
	blackQueen  = "\u265B"
	blackRook   = "\u265C"
	blackBishop = "\u265D"
	blackKnight = "\u265E"
	blackPawn   = "\u265F"
)

const (
	fgBlack = iota + 30
)

// Background text colors
const (
	bgBlack = iota + 40
	bgRed
	bgGreen
	bgYellow
	bgBlue
	bgMagenta
	bgCyan
	bgWhite
)

// Background Hi-Intensity text colors
const (
	bgHiBlack = iota + 100
	bgHiRed
	bgHiGreen
	bgHiYellow
	bgHiBlue
	bgHiMagenta
	bgHiCyan
	bgHiWhite
)

var chessSymbols = [2][7]string{
	{" ", whitePawn, whiteKnight, whiteBishop, whiteRook, whiteQueen, whiteKing},
	{" ", blackPawn, blackKnight, blackBishop, blackRook, blackQueen, blackKing},
}

func PrintPosition(p *common.Position) {
	for i := 0; i < 64; i++ {
		sq := common.FlipSquare(i)
		piece, side := p.GetPieceTypeAndSide(sq)
		fmt.Print(pieceString(piece, side, common.IsDarkSquare(sq)))
		if common.File(sq) == common.FileH {
			fmt.Println()
		}
	}
}

func pieceString(piece int, side, darkSquare bool) string {
	var s string
	if side {
		s = chessSymbols[0][piece]
	} else {
		s = chessSymbols[1][piece]
	}
	s += " "
	const fgColor = fgBlack
	var bgColor int
	if darkSquare {
		bgColor = bgWhite
	} else {
		bgColor = bgHiWhite
	}
	const escape = "\x1b"
	const reset = 0
	return fmt.Sprintf("%s[%s;%sm%s%s[%dm",
		escape, strconv.Itoa(fgColor), strconv.Itoa(bgColor), s, escape, reset)
}
