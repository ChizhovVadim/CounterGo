package shell

import (
	"counter/engine"
	"fmt"
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

var chessSymbols = [2][7]string{
	{" ", whitePawn, whiteKnight, whiteBishop, whiteRook, whiteQueen, whiteKing},
	{" ", blackPawn, blackKnight, blackBishop, blackRook, blackQueen, blackKing},
}

func PrintPosition(p *engine.Position) {
	for i := 0; i < 64; i++ {
		sq := engine.FlipSquare(i)
		piece, side := p.GetPieceTypeAndSide(sq)
		if side {
			fmt.Print(chessSymbols[0][piece])
		} else {
			fmt.Print(chessSymbols[1][piece])
		}
		fmt.Print(" ")
		if engine.File(sq) == engine.FileH {
			fmt.Println()
		}
	}
}
