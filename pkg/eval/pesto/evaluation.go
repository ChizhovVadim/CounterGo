package eval

import (
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	pawnValue = 100
)

const (
	minorPhase = 4
	rookPhase  = 6
	queenPhase = 12
	totalPhase = 2 * (4*minorPhase + 2*rookPhase + queenPhase)
)

const (
	darkSquares = uint64(0xAA55AA55AA55AA55)
)

type EvaluationService struct {
	Weights
	pieceCount [2][King + 1]int
	force      [2]int
}

func NewEvaluationService() *EvaluationService {
	var es = &EvaluationService{}
	es.Weights.init()
	return es
}

func (e *EvaluationService) Evaluate(p *Position) int {
	var (
		x     uint64
		sq    int
		piece int
		s     Score
	)

	for piece = Pawn; piece <= King; piece++ {
		e.pieceCount[SideWhite][piece] = 0
		e.pieceCount[SideBlack][piece] = 0
	}

	for x = p.White; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		piece = p.WhatPiece(sq)
		s += e.PST[SideWhite][piece][sq]
		e.pieceCount[SideWhite][piece]++
	}

	for x = p.Black; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		piece = p.WhatPiece(sq)
		s += e.PST[SideBlack][piece][sq]
		e.pieceCount[SideBlack][piece]++
	}

	e.force[SideWhite] = minorPhase*(e.pieceCount[SideWhite][Knight]+e.pieceCount[SideWhite][Bishop]) +
		rookPhase*e.pieceCount[SideWhite][Rook] + queenPhase*e.pieceCount[SideWhite][Queen]
	e.force[SideBlack] = minorPhase*(e.pieceCount[SideBlack][Knight]+e.pieceCount[SideBlack][Bishop]) +
		rookPhase*e.pieceCount[SideBlack][Rook] + queenPhase*e.pieceCount[SideBlack][Queen]

	if e.pieceCount[SideWhite][Bishop] >= 2 {
		s += e.BishopPairMaterial
	}
	if e.pieceCount[SideBlack][Bishop] >= 2 {
		s -= e.BishopPairMaterial
	}

	// mix score

	var phase = e.force[SideWhite] + e.force[SideBlack]
	if phase > totalPhase {
		phase = totalPhase
	}

	var result = (int(s.Middle())*phase + int(s.End())*(totalPhase-phase)) / totalPhase

	var ocb = e.force[SideWhite] == minorPhase &&
		e.force[SideBlack] == minorPhase &&
		(p.Bishops&darkSquares) != 0 &&
		(p.Bishops & ^darkSquares) != 0

	if result > 0 {
		result = result * computeFactor(e, SideWhite, ocb) / scaleNormal
	} else {
		result = result * computeFactor(e, SideBlack, ocb) / scaleNormal
	}

	if !p.WhiteMove {
		result = -result
	}

	return result
}

const (
	scaleDraw   = 0
	scaleHard   = 1
	scaleNormal = 2
)

func computeFactor(e *EvaluationService, side int, ocb bool) int {
	if e.force[side] >= queenPhase+rookPhase {
		return scaleNormal
	}
	if e.pieceCount[side][Pawn] == 0 {
		if e.force[side] <= minorPhase {
			return scaleHard
		}
		if e.force[side] == 2*minorPhase && e.pieceCount[side][Knight] == 2 && e.pieceCount[side^1][Pawn] == 0 {
			return scaleHard
		}
		if e.force[side]-e.force[side^1] <= minorPhase {
			return scaleHard
		}
	} else if e.pieceCount[side][Pawn] == 1 {
		if e.force[side] <= minorPhase && e.pieceCount[side^1][Knight]+e.pieceCount[side^1][Bishop] != 0 {
			return scaleHard
		}
		if e.force[side] == e.force[side^1] && e.pieceCount[side^1][Knight]+e.pieceCount[side^1][Bishop] != 0 {
			return scaleHard
		}
	} else if ocb && e.pieceCount[side][Pawn]-e.pieceCount[side^1][Pawn] <= 2 {
		return scaleHard
	}
	return scaleNormal
}

func sameColorSquares(sq int) uint64 {
	if IsDarkSquare(sq) {
		return darkSquares
	}
	return ^darkSquares
}
