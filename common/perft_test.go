package common

import (
	"testing"
)

//https://chessprogramming.wikispaces.com/Perft+Results
func TestPerft(t *testing.T) {
	var tests = []struct {
		fen   string
		depth int
		nodes int
	}{
		{
			fen:   InitialPositionFen,
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
		var p = NewPositionFromFEN(test.fen)
		var nodes = Perft(p, test.depth)
		if nodes != test.nodes {
			t.Error(i, test, nodes)
		}
	}
}

func Perft(p *Position, depth int) int {
	var result = 0
	var buffer [MAX_MOVES]Move
	var child Position
	for _, move := range GenerateMoves(buffer[:], p) {
		if p.MakeMove(move, &child) {
			if depth > 1 {
				result += Perft(&child, depth-1)
			} else {
				result++
			}
		}
	}
	return result
}
