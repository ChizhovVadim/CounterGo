package engine

import (
	"errors"
	"testing"
)

func TestSEE(t *testing.T) {
	var tests = []struct {
		position string
		move     string
	}{
		{"1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - -", "e1e5"},
		{"1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - -", "d3e5"},
		{"4k3/8/2n5/4b3/8/3N4/8/4K3 w - - 0 1", "d3e5"},
		{"5kn1/7P/8/8/8/8/8/4K3 w - - 0 1", "h7g8q"},

		{"8/5r1p/5k2/4R3/p1p1KP2/P7/1P1p3P/8 w - - 2 2", "e5f5"},
		{"8/8/8/1p2q3/1P2rkp1/2P5/5K1Q/8 b - - 6 4", "g4g3"},
	}
	for _, test := range tests {
		var p = NewPositionFromFEN(test.position)
		var m = parseMove(p, test.move)
		if m == MoveEmpty {
			t.Error("Wrong move", test)
			continue
		}
		var directSEE, err = SEE_GE_Direct(p, m)
		if err != nil {
			t.Error("Wrong move", test)
			continue
		}
		if see := SEE_GE(p, m); see != directSEE {
			t.Error("Bad SEE", directSEE, see, test)
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

func BasicMaterial(p *Position) int {
	var score = 0
	score += PopCount(p.Pawns&p.White) - PopCount(p.Pawns&p.Black)
	score += 4 * (PopCount(p.Knights&p.White) - PopCount(p.Knights&p.Black))
	score += 4 * (PopCount(p.Bishops&p.White) - PopCount(p.Bishops&p.Black))
	score += 6 * (PopCount(p.Rooks&p.White) - PopCount(p.Rooks&p.Black))
	score += 12 * (PopCount(p.Queens&p.White) - PopCount(p.Queens&p.Black))
	if !p.WhiteMove {
		score = -score
	}
	return score
}

func SEE_GE_Direct(p *Position, move Move) (goodMove bool, err error) {
	var eval = BasicMaterial(p)
	var child = &Position{}
	if !p.MakeMove(move, child) {
		return false, errors.New("wrong move")
	}
	var score = -SearchQSRecapture(child, -VALUE_INFINITE, -(eval - 1), move.To())
	return score >= eval, nil
}

func SearchQSRecapture(p *Position, alpha, beta, square int) int {
	var eval = BasicMaterial(p)
	if eval > alpha {
		alpha = eval
		if eval >= beta {
			return eval
		}
	}
	var ml MoveList
	ml.GenerateCaptures(p, false)
	var child = &Position{}
	for i := 0; i < ml.Count; i++ {
		var move = ml.Items[i].Move
		if move.To() == square && p.MakeMove(move, child) {
			var score = -SearchQSRecapture(child, -beta, -alpha, square)
			if score > alpha {
				alpha = score
				if score >= beta {
					return score
				}
			}
		}
	}
	return alpha
}
