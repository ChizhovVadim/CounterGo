package train

import (
	"math/rand"

	"github.com/ChizhovVadim/CounterGo/internal/ml"
)

type Model struct {
	layer1 *Layer
	layer2 *Layer
	cost   ml.IModelCost
}

func NewModelNN(rnd *rand.Rand) *Model {
	var hiddenSize = 512
	return &Model{
		layer1: NewLayer(
			768,
			make([]Neuron, hiddenSize),
			&ml.ReLuActivation{}).
			InitWeightsReLU(rnd), // ненулевых входных признаков не более 32 (кол во фигур на доске)
		layer2: NewLayer(
			hiddenSize,
			make([]Neuron, 1),
			&ml.SigmoidActivation{}).
			InitWeightsSigmoid(rnd),
		cost: &ml.MSECost{},
	}
}

func (m *Model) ThreadCopy() *Model {
	return &Model{
		layer1: m.layer1.ThreadCopy(),
		layer2: m.layer2.ThreadCopy(),
		cost:   m.cost,
	}
}

// TODO input FeatureSet
func (m *Model) forward(sample *Sample) float64 {
	m.layer1.Forward(nil, sample.Features)
	m.layer2.Forward(m.layer1.outputs, nil)
	return m.layer2.outputs[0].Activation
}

func (m *Model) CalcCost(sample *Sample) float64 {
	predicted := m.forward(sample)
	return m.cost.Cost(predicted, float64(sample.Target))
}

func (m *Model) Train(sample *Sample) {
	predicted := m.forward(sample)
	m.layer2.outputs[0].Error = m.cost.CostPrime(predicted, float64(sample.Target))
	// back propagation
	m.layer2.Backward(m.layer1.outputs, nil)
	m.layer1.Backward(nil, sample.Features)
}

func (m *Model) AddGradients(mainModel *Model) {
	if m == mainModel {
		return
	}
	m.layer1.AddGradients(mainModel.layer1)
	m.layer2.AddGradients(mainModel.layer2)
}

func (m *Model) ApplyGradients() {
	m.layer1.ApplyGradients()
	m.layer2.ApplyGradients()
}

func (m *Model) Load(filepath string) error {
	var n = LoadNetwork(filepath)
	m.layer1.weights = n.Weights[0]
	m.layer1.biases = n.Biases[0]
	m.layer2.weights = n.Weights[1]
	m.layer2.biases = n.Biases[1]
	return nil
}

func (m *Model) Save(filepath string) error {
	var n = &Network{
		Topology: Topology{
			Inputs:        768,
			HiddenNeurons: []uint32{512},
			Outputs:       1,
		},
		Weights: []ml.Matrix{m.layer1.weights, m.layer2.weights},
		Biases:  []ml.Matrix{m.layer1.biases, m.layer2.biases},
	}
	return n.Save(filepath)
}
