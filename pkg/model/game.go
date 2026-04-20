package model

import (
	"maps"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Game struct {
	Position common.Position
	repeats  map[uint64]int // включая текущую позицию
}

func NewGame(startFen string) (Game, error) {
	if startFen == "" {
		startFen = common.InitialPositionFen
	}
	var startPos, err = common.NewPositionFromFEN(startFen)
	if err != nil {
		return Game{}, err
	}
	return Game{
		Position: startPos,
		repeats:  map[uint64]int{startPos.Key: 1},
	}, nil
}

func (g *Game) Clone() Game {
	return Game{
		Position: g.Position,
		repeats:  maps.Clone(g.repeats),
	}
}

func (g *Game) MakeMove(move common.Move) bool {
	var buffer [common.MaxMoves]common.OrderedMove
	var child common.Position
	var ml = g.Position.GenerateMoves(buffer[:])
	for i := range ml {
		var mv = ml[i].Move
		if mv != move {
			continue
		}
		if !g.Position.MakeMove(mv, &child) {
			return false
		}
		g.Position = child
		if mv.MovingPiece() == common.Pawn ||
			mv.CapturedPiece() != common.Empty {
			clear(g.repeats)
		}
		g.repeats[g.Position.Key] += 1
		return true
	}
	return false
}

func (g *Game) IsTwoTimeRepeats(key uint64) bool {
	return g.repeats[key] >= 2
}
