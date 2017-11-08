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

func basicMaterial(p *Position) int {
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
	var eval = basicMaterial(p)
	var child = &Position{}
	if !p.MakeMove(move, child) {
		return false, errors.New("wrong move")
	}
	var score = -searchSEE(child, -VALUE_INFINITE, -(eval - 1), move.To())
	return score >= eval, nil
}

func searchSEE(p *Position, alpha, beta, square int) int {
	var eval = basicMaterial(p)
	if eval > alpha {
		alpha = eval
		if eval >= beta {
			return eval
		}
	}
	var ml MoveList
	ml.GenerateCaptures(p, false)
	var child = &Position{}
	var move = lvaRecapture(p, child, &ml, square)
	if move != MoveEmpty &&
		p.MakeMove(move, child) {
		var score = -searchSEE(child, -beta, -alpha, square)
		if score > alpha {
			alpha = score
			if score >= beta {
				return score
			}
		}
	}
	return alpha
}

func lvaRecapture(p, child *Position, ml *MoveList, square int) Move {
	var piece = King + 1
	var bestMove = MoveEmpty
	for i := 0; i < ml.Count; i++ {
		var move = ml.Items[i].Move
		if move.To() == square &&
			move.MovingPiece() < piece &&
			p.MakeMove(move, child) {
			bestMove = move
			piece = move.MovingPiece()
		}
	}
	return bestMove
}

func TestTransTable(t *testing.T) {
	var transTable = NewTransTable(1)
	var p = NewPositionFromFEN(InitialPositionFen)
	var depth = 5
	var score = 5
	var bound = Lower
	var move = MoveEmpty
	transTable.Update(p, depth, score, bound, move)
	var ttDepth, ttScore, ttBound, ttMove, ttOk = transTable.Read(p)
	if !ttOk || depth != ttDepth || score != ttScore ||
		bound != ttBound || move != ttMove {
		t.Error()
	}
}

const (
	BB_A1 = uint64(1) << iota
	BB_B1
	BB_C1
	BB_D1
	BB_E1
	BB_F1
	BB_G1
	BB_H1
	BB_A2
	BB_B2
	BB_C2
	BB_D2
	BB_E2
	BB_F2
	BB_G2
	BB_H2
	BB_A3
	BB_B3
	BB_C3
	BB_D3
	BB_E3
	BB_F3
	BB_G3
	BB_H3
	BB_A4
	BB_B4
	BB_C4
	BB_D4
	BB_E4
	BB_F4
	BB_G4
	BB_H4
	BB_A5
	BB_B5
	BB_C5
	BB_D5
	BB_E5
	BB_F5
	BB_G5
	BB_H5
	BB_A6
	BB_B6
	BB_C6
	BB_D6
	BB_E6
	BB_F6
	BB_G6
	BB_H6
	BB_A7
	BB_B7
	BB_C7
	BB_D7
	BB_E7
	BB_F7
	BB_G7
	BB_H7
	BB_A8
	BB_B8
	BB_C8
	BB_D8
	BB_E8
	BB_F8
	BB_G8
	BB_H8
)

func TestPawns(t *testing.T) {
	var tests = []struct {
		fen           string
		doubledPawns  uint64
		isolatedPanws uint64
		passedPawns   uint64
	}{
		{
			fen:           "1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - -",
			doubledPawns:  0,
			isolatedPanws: BB_E5 | BB_H7,
			passedPawns:   BB_E5,
		},
	}
	for _, test := range tests {
		var p = NewPositionFromFEN(test.fen)
		if v := GetDoubledPawns(p.Pawns&p.White) | GetDoubledPawns(p.Pawns&p.Black); v != test.doubledPawns {
			t.Error("doubled", test.fen, bitboardToString(v))
		}
		if v := GetIsolatedPawns(p.Pawns&p.White) | GetIsolatedPawns(p.Pawns&p.Black); v != test.isolatedPanws {
			t.Error("isolated", test.fen, bitboardToString(v))
		}
		if v := GetWhitePassedPawns(p) | GetBlackPassedPawns(p); v != test.passedPawns {
			t.Error("passed", test.fen, bitboardToString(v))
		}
	}
}

func bitboardToString(bb uint64) string {
	var s = ""
	for b := bb; b != 0; b &= b - 1 {
		var sq = FirstOne(b)
		if len(s) > 0 {
			s += ","
		}
		s += SquareName(sq)
	}
	return s
}
