package tuner

import (
	"fmt"
	"math"

	"github.com/ChizhovVadim/CounterGo/internal/ml"
)

// Начальные веса не задаем (линейная функция)
// смещения не используем (симметричная функция с точки зрения белых-черных)
type Model struct {
	activationFn ml.IActivationFn
	weights      ml.Matrix
	wGradients   ml.Gradients
	cost         ml.IModelCost
}

func NewModelHCE(
	inputSize int,
) *Model {
	return &Model{
		activationFn: &ml.SigmoidActivation{},
		weights:      ml.NewMatrix(2, inputSize),
		wGradients:   ml.NewGradients(2, inputSize),
		cost:         &ml.MSECost{},
	}
}

func (m *Model) ApplyGradients() {
	m.wGradients.Apply(&m.weights)
}

func (m *Model) CalcCost(sample *Sample) float64 {
	var cost float64
	m.work(sample, false, &cost)
	return cost
}

func (m *Model) Train(sample *Sample) {
	var cost float64
	m.work(sample, true, &cost)
}

func (m *Model) work(sample *Sample, train bool, cost *float64) {
	const (
		Opening = 0
		Endgame = 1
	)
	var mg, eg float64
	for _, input := range sample.Features {
		var inputIndex = int(input.Index)
		var inputValue = float64(input.Value)
		mg += m.weights.Get(Opening, inputIndex) * inputValue
		eg += m.weights.Get(Endgame, inputIndex) * inputValue
	}
	var phase = float64(sample.MgPhase)
	var mix = phase*mg + (1-phase)*eg
	var strongSideScale float64
	if mix > 0 {
		strongSideScale = float64(sample.WhiteStrongScale)
	} else {
		strongSideScale = float64(sample.BlackStrongScale)
	}
	var x = mix * strongSideScale
	var predicted = m.activationFn.Sigma(x)
	if !train {
		*cost = m.cost.Cost(predicted, float64(sample.Target))
		return
	}
	// back propagation
	var outputGradient = m.cost.CostPrime(predicted, float64(sample.Target)) *
		m.activationFn.SigmaPrime(x) *
		strongSideScale
	for _, input := range sample.Features {
		var inputIndex = int(input.Index)
		var inputValue = float64(input.Value)
		m.wGradients.Add(Opening, inputIndex, inputValue*phase*outputGradient)
		m.wGradients.Add(Endgame, inputIndex, inputValue*(1-phase)*outputGradient)
	}
}

func (m *Model) Print() {
	const ScaleEval = 10_000
	var weights = m.weights.Data
	var wInt = make([]int, len(weights))
	for i := range wInt {
		wInt[i] = int(math.Round(ScaleEval * weights[i]))
	}
	//fmt.Printf("%#v\n", weights)
	fmt.Printf("%#v\n", wInt)
}

func (m *Model) ThreadCopy() *Model {
	return &Model{
		activationFn: m.activationFn,
		weights:      m.weights,
		wGradients:   ml.NewGradients(m.wGradients.Rows, m.weights.Cols),
		cost:         m.cost,
	}
}

func (m *Model) AddGradients(mainModel *Model) {
	if m == mainModel {
		return
	}
	m.wGradients.AddTo(&mainModel.wGradients)
}
