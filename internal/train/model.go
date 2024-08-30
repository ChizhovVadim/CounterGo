package train

import (
	"encoding/binary"
	"io"
	"math"
	"math/rand"
	"os"

	"github.com/ChizhovVadim/CounterGo/internal/ml"
)

type Model struct {
	layer1 *Layer
	layer2 *Layer
}

func NewModel(inputSize, hiddenSize int) *Model {
	return &Model{
		layer1: NewLayer(
			inputSize,
			make([]Neuron, hiddenSize),
			&ml.ReLuActivation{}),
		layer2: NewLayer(
			hiddenSize,
			make([]Neuron, 1),
			&ml.SigmoidActivation{}),
	}
}

func (m *Model) InitWeights(rnd *rand.Rand) {
	m.layer1.InitWeightsReLU(rnd) // ненулевых входных признаков не более 32 (кол во фигур на доске)
	m.layer2.InitWeightsSigmoid(rnd)
}

func (m *Model) Forward(input *Input) float64 {
	m.layer1.Forward(nil, input.Features)
	m.layer2.Forward(m.layer1.outputs, nil)
	return m.layer2.outputs[0].Activation
}

func (m *Model) Train(sample *Sample, cost ml.IModelCost) {
	predicted := m.Forward(&sample.input)
	m.layer2.outputs[0].Error = cost.CostPrime(predicted, float64(sample.target))
	// back propagation
	m.layer2.Backward(m.layer1.outputs, nil)
	m.layer1.Backward(nil, sample.input.Features)
}

func (m *Model) ApplyGradients() {
	m.layer1.ApplyGradients()
	m.layer2.ApplyGradients()
}

func (m *Model) Clone() IModel {
	return &Model{
		layer1: m.layer1.ThreadCopy(),
		layer2: m.layer2.ThreadCopy(),
	}
}

func (m *Model) AddGradients(abstractMainModel IModel) {
	var mainModel = abstractMainModel.(*Model)
	if m == mainModel {
		return
	}
	m.layer1.AddGradients(mainModel.layer1)
	m.layer2.AddGradients(mainModel.layer2)
}

func (m *Model) LoadWeights(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var data = [...][]float64{
		m.layer1.weights.Data,
		m.layer1.biases.Data,
		m.layer2.weights.Data,
		m.layer2.biases.Data,
	}
	for i := range data {
		var err = readSlice(f, data[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) SaveWeights(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var data = [...][]float64{
		m.layer1.weights.Data,
		m.layer1.biases.Data,
		m.layer2.weights.Data,
		m.layer2.biases.Data,
	}
	for i := range data {
		var err = writeSlice(f, data[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func readSlice(f io.Reader, data []float64) error {
	var buf [4]byte
	for i := range data {
		_, err := io.ReadFull(f, buf[:])
		if err != nil {
			return err
		}
		var val = math.Float32frombits(binary.LittleEndian.Uint32(buf[:]))
		data[i] = float64(val)
	}
	return nil
}

func writeSlice(f io.Writer, data []float64) error {
	var buf [4]byte
	for i := range data {
		var val = float32(data[i])
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(val))
		_, err := f.Write(buf[:])
		if err != nil {
			return err
		}
	}
	return nil
}
