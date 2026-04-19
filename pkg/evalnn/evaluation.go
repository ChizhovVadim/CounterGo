package evalnn

import (
	"math"

	"github.com/ChizhovVadim/CounterGo/internal/simd"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	Add    = 1
	Remove = -Add
)

const MaxHeight = 128

type EvaluationService struct {
	weights            *Weights
	scale              float32
	accumulators       [MaxHeight]Accumulator
	currentAccumulator int
	updates            Updates
	hiddenActivation   [HiddenSize]float32
}

type Accumulator struct {
	hiddenOutputs [HiddenSize]float32
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

func NewEvaluationService(weights *Weights, scale float32) *EvaluationService {
	return &EvaluationService{
		weights: weights,
		scale:   scale,
	}
}

func (e *EvaluationService) Init(p *common.Position) {
	input := make([]int, 0, 32)

	for b := p.AllPieces(); b != 0; b &= b - 1 {
		var sq = common.FirstOne(b)
		piece, side := p.GetPieceTypeAndSide(sq)
		if piece != common.Empty {
			input = append(input, int(calculateNetInputIndex(side, piece, sq)))
		}
	}

	e.currentAccumulator = 0
	var acc = &e.accumulators[e.currentAccumulator]

	hiddenOutputs := acc.hiddenOutputs[:]
	for i := range hiddenOutputs {
		hiddenOutputs[i] = e.weights.hiddenBiases[i]
	}
	for _, i := range input {
		for j := range hiddenOutputs {
			hiddenOutputs[j] += e.weights.hiddenWeights[i*HiddenSize+j]
		}
	}
}

func (e *EvaluationService) MakeMove(p *common.Position, m common.Move) {
	calculateUpdates(p, m, &e.updates)

	e.currentAccumulator += 1
	var hiddenOutputs = e.accumulators[e.currentAccumulator].hiddenOutputs[:]
	copy(hiddenOutputs, e.accumulators[e.currentAccumulator-1].hiddenOutputs[:])

	for i := range e.updates.Size {
		var index = int(e.updates.Indices[i]) * HiddenSize
		var coeff = float32(e.updates.Coeffs[i])
		simd.AddScaled(hiddenOutputs, coeff, e.weights.hiddenWeights[index:index+HiddenSize])
	}
}

func (e *EvaluationService) UnmakeMove() {
	e.currentAccumulator -= 1
}

func (e *EvaluationService) EvaluateQuick(p *common.Position) int {
	var output = int(e.scale * e.quickFeed())
	const MaxEval = 15_000
	output = max(-MaxEval, min(MaxEval, output))

	var npMaterial = 4*common.PopCount(p.Knights|p.Bishops) + 6*common.PopCount(p.Rooks) + 12*common.PopCount(p.Queens)
	output = output * (160 + npMaterial) / 160
	output = output * (200 - p.Rule50) / 200

	if !p.WhiteMove {
		output = -output
	}
	return output
}

func (e *EvaluationService) EvaluateProb(p *common.Position) float64 {
	e.Init(p)
	var output = float64(e.quickFeed())
	return sigmoid(output)
}

func (e *EvaluationService) quickFeed() float32 {
	simd.Relu(e.hiddenActivation[:], e.accumulators[e.currentAccumulator].hiddenOutputs[:])
	return simd.DotProduct(e.hiddenActivation[:], e.weights.outputWeights[:]) + e.weights.outputBias
}

func calculateNetInputIndex(whiteSide bool, pieceType, square int) int16 {
	var piece12 = pieceType - common.Pawn
	if !whiteSide {
		piece12 += 6
	}
	return int16(square ^ piece12<<6)
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}
