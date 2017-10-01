package shell

import (
	"fmt"
	"strconv"

	"github.com/ChizhovVadim/CounterGo/engine"
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
	FgBlack = iota + 30
)

// Background text colors
const (
	BgBlack = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// Background Hi-Intensity text colors
const (
	BgHiBlack = iota + 100
	BgHiRed
	BgHiGreen
	BgHiYellow
	BgHiBlue
	BgHiMagenta
	BgHiCyan
	BgHiWhite
)

var chessSymbols = [2][7]string{
	{" ", whitePawn, whiteKnight, whiteBishop, whiteRook, whiteQueen, whiteKing},
	{" ", blackPawn, blackKnight, blackBishop, blackRook, blackQueen, blackKing},
}

func PrintPosition(p *engine.Position) {
	for i := 0; i < 64; i++ {
		sq := engine.FlipSquare(i)
		piece, side := p.GetPieceTypeAndSide(sq)
		fmt.Print(PieceString(piece, side, engine.IsDarkSquare(sq)))
		if engine.File(sq) == engine.FileH {
			fmt.Println()
		}
	}
}

func PieceString(piece int, side, darkSquare bool) string {
	var s string
	if side {
		s = chessSymbols[0][piece]
	} else {
		s = chessSymbols[1][piece]
	}
	s += " "
	const fgColor = FgBlack
	var bgColor int
	if darkSquare {
		bgColor = BgWhite
	} else {
		bgColor = BgHiWhite
	}
	const escape = "\x1b"
	const reset = 0
	return fmt.Sprintf("%s[%s;%sm%s%s[%dm",
		escape, strconv.Itoa(fgColor), strconv.Itoa(bgColor), s, escape, reset)
}
