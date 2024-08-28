package eval

import (
	"github.com/ChizhovVadim/CounterGo/internal/math"
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	InputSize  = 64 * 12
	HiddenSize = 512
)

const (
	Add    = 1
	Remove = -Add
)

const MaxHeight = 128

type EvaluationService struct {
	*Weights
	updates       Updates
	hiddenOutputs [MaxHeight][HiddenSize]float32
	currentHidden int
}

type Weights struct {
	HiddenWeights [InputSize * HiddenSize]float32
	HiddenBiases  [HiddenSize]float32
	OutputWeights [HiddenSize]float32
	OutputBias    float32
}

type Updates struct {
	Indices [8]int16
	Coeffs  [8]int8
	Size    int
}

func (u *Updates) Add(index int16, coeff int8) {
	u.Indices[u.Size] = index
	u.Coeffs[u.Size] = coeff
	u.Size++
}

func NewEvaluationService(weights *Weights) *EvaluationService {
	var es = &EvaluationService{}
	es.Weights = weights
	return es
}

func (e *EvaluationService) EvaluateQuick(p *Position) int {
	output := int(e.QuickFeed())
	const MaxEval = 15_000
	output = Max(-MaxEval, Min(MaxEval, output))
	var npMaterial = 4*PopCount(p.Knights|p.Bishops) + 6*PopCount(p.Rooks) + 12*PopCount(p.Queens)
	output = output * (160 + npMaterial) / 160
	output = output * (200 - p.Rule50) / 200
	if !p.WhiteMove {
		output = -output
	}
	return output
}

func (e *EvaluationService) Evaluate(p *Position) int {
	e.Init(p)
	return e.EvaluateQuick(p)
}

func (e *EvaluationService) Init(p *Position) {
	input := make([]int, 0, 33)

	for sq := 0; sq < 64; sq++ {
		piece, side := p.GetPieceTypeAndSide(sq)
		if piece != Empty {
			input = append(input, int(calculateNetInputIndex(side, piece, sq)))
		}
	}

	e.currentHidden = 0
	hiddenOutputs := e.hiddenOutputs[e.currentHidden][:]

	for i := range hiddenOutputs {
		hiddenOutputs[i] = e.HiddenBiases[i]
	}

	for _, i := range input {
		for j := range hiddenOutputs {
			hiddenOutputs[j] += e.HiddenWeights[i*HiddenSize+j]
		}
	}
}

func calculateNetInputIndex(whiteSide bool, pieceType, square int) int16 {
	var piece12 = pieceType - Pawn
	if !whiteSide {
		piece12 += 6
	}
	return int16(square ^ piece12<<6)
}

func (e *EvaluationService) MakeMove(p *Position, m Move) {
	e.updates.Size = 0

	// MakeNullMove
	if m == MoveEmpty {
		e.UpdateHidden()
		return
	}

	var from, to, movingPiece, capturedPiece, epCapSq, promotionPt, isCastling = unpackMove(p, m)

	e.updates.Add(calculateNetInputIndex(p.WhiteMove, movingPiece, from), Remove)

	if capturedPiece != Empty {
		var capSq = to
		if epCapSq != SquareNone {
			capSq = epCapSq
		}
		e.updates.Add(calculateNetInputIndex(!p.WhiteMove, capturedPiece, capSq), Remove)
	}

	var pieceAfterMove = movingPiece
	if promotionPt != Empty {
		pieceAfterMove = promotionPt
	}
	e.updates.Add(calculateNetInputIndex(p.WhiteMove, pieceAfterMove, to), Add)

	if isCastling {
		var rookRemoveSq, rookAddSq int
		if p.WhiteMove {
			if to == SquareG1 {
				rookRemoveSq = SquareH1
				rookAddSq = SquareF1
			} else {
				rookRemoveSq = SquareA1
				rookAddSq = SquareD1
			}
		} else {
			if to == SquareG8 {
				rookRemoveSq = SquareH8
				rookAddSq = SquareF8
			} else {
				rookRemoveSq = SquareA8
				rookAddSq = SquareD8
			}
		}

		e.updates.Add(calculateNetInputIndex(p.WhiteMove, Rook, rookRemoveSq), Remove)
		e.updates.Add(calculateNetInputIndex(p.WhiteMove, Rook, rookAddSq), Add)
	}

	e.UpdateHidden()
}

func (e *EvaluationService) UnmakeMove() {
	e.currentHidden--
}

func unpackMove(p *Position, m Move) (from, to, movingPiece, capturedPiece, epCapSq, promotionPt int, isCastling bool) {
	from = m.From()
	to = m.To()
	movingPiece = m.MovingPiece()
	capturedPiece = m.CapturedPiece()
	promotionPt = m.Promotion()
	epCapSq = SquareNone
	if movingPiece == King {
		if p.WhiteMove {
			if from == SquareE1 && (to == SquareG1 || to == SquareC1) {
				isCastling = true
			}
		} else {
			if from == SquareE8 && (to == SquareG8 || to == SquareC8) {
				isCastling = true
			}
		}
	} else if movingPiece == Pawn {
		if to == p.EpSquare {
			if p.WhiteMove {
				epCapSq = to - 8
			} else {
				epCapSq = to + 8
			}
		}
	}
	return
}

func (e *EvaluationService) EvaluateProb(p *Position) float64 {
	var centipawns = e.Evaluate(p)
	if !p.WhiteMove {
		centipawns = -centipawns
	}
	return math.Sigmoid(3.5 / 512 * float64(centipawns))
}
