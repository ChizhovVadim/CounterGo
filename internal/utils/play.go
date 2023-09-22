package utils

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type IEngine interface {
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

func PlayCli(engine IEngine) error {
	var game = newGame()
	game.Print()
	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var commandLine = scanner.Text()
		if commandLine == "quit" {
			break
		}
		if !game.MakeMoveLAN(commandLine) {
			fmt.Println("bad move")
			continue
		}
		game.Print()
		var si = engine.Search(context.Background(), common.SearchParams{
			Positions: game.positions,
			Limits: common.LimitsType{
				MoveTime: 3000,
			},
		})
		if len(si.MainLine) == 0 {
			return fmt.Errorf("bad move")
		}
		var bestMove = si.MainLine[0]
		fmt.Println(bestMove.String())
		if !game.MakeMove(bestMove) {
			return fmt.Errorf("bad move %v", bestMove)
		}
		game.Print()
	}
	return nil
}

type game struct {
	positions []common.Position
}

func newGame() *game {
	var pos, _ = common.NewPositionFromFEN(common.InitialPositionFen)
	return &game{
		positions: []common.Position{pos},
	}
}

func (g *game) Print() {
	var curPos = &g.positions[len(g.positions)-1]
	for i := 0; i < 64; i++ {
		sq := common.FlipSquare(i)
		piece, side := curPos.GetPieceTypeAndSide(sq)
		fmt.Print(pieceString(piece, side, common.IsDarkSquare(sq)))
		if common.File(sq) == common.FileH {
			fmt.Println()
		}
	}
}

func (g *game) MakeMoveLAN(smove string) bool {
	var curPos = &g.positions[len(g.positions)-1]
	var child, ok = curPos.MakeMoveLAN(smove)
	if !ok {
		return false
	}
	g.positions = append(g.positions, child)
	return true
}

func (g *game) MakeMove(move common.Move) bool {
	var child common.Position
	if !g.positions[len(g.positions)-1].MakeMove(move, &child) {
		return false
	}
	g.positions = append(g.positions, child)
	return true
}

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
