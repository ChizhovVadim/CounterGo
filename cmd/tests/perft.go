package main

import (
	"fmt"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func perftHandler(args []string) error {
	var tests = []struct {
		fen   string
		depth int
		nodes int
	}{
		{
			fen:   common.InitialPositionFen,
			depth: 6,
			nodes: 119060324,
		},
		{
			fen:   "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -",
			depth: 5,
			nodes: 193690690,
		},
		{
			fen:   "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -",
			depth: 7,
			nodes: 178633661,
		},
		{
			fen:   "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
			depth: 5,
			nodes: 15833292,
		},
		{
			fen:   "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
			depth: 5,
			nodes: 89941194,
		},
		{
			fen:   "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
			depth: 5,
			nodes: 164075551,
		},
	}
	for i, test := range tests {
		var p, err = common.NewPositionFromFEN(test.fen)
		if err != nil {
			return err
		}
		var nodes = perft(&p, test.depth)
		if nodes != test.nodes {
			return fmt.Errorf("perft failed %v %v %v",
				i, test, nodes)
		}
	}
	return nil
}

func perft(p *common.Position, depth int) int {
	var result = 0
	var buffer [common.MaxMoves]common.OrderedMove
	var child common.Position
	var ml = p.GenerateMoves(buffer[:])
	for i := range ml {
		var move = ml[i].Move
		if p.MakeMove(move, &child) {
			if depth > 1 {
				result += perft(&child, depth-1)
			} else {
				result++
			}
		}
	}
	return result
}
