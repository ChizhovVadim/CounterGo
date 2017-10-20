package engine

import "testing"

func TestSEE(t *testing.T) {
	var tests = []struct {
		position string
		move     string
		see      int
	}{
		{"1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - -", "e1e5", 1},
		{"1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - -", "d3e5", -3},
	}
	for _, test := range tests {
		var p = NewPositionFromFEN(test.position)
		var m = parseMove(p, test.move)
		if m == MoveEmpty {
			t.Error("Bad move", test)
			continue
		}
		if see := SEE(p, m); see != test.see {
			t.Error("Bad SEE", see, test)
		}
	}
}

func parseMove(p *Position, smove string) Move {
	var move = ParseMove(smove)
	var ml MoveList
	ml.GenerateMoves(p)
	ml.FilterLegalMoves(p)
	for i := 0; i < ml.Count; i++ {
		var x = ml.Items[i].Move
		if move.From() == x.From() && move.To() == x.To() && move.Promotion() == x.Promotion() {
			return x
		}
	}
	return MoveEmpty
}
