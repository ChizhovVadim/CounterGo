package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"path/filepath"
	"sync"
	"sync/atomic"
)

var (
	SigmoidScale = 3.5 / 512
	LearningRate = 0.01
	BatchSize    = 16384
)

type Neuron struct {
	A, E, Prime float64
}

type ThreadData struct {
	wGradients []Matrix
	bGradients []Matrix
	neurons    [][]Neuron
}

type Trainer struct {
	topology    []int
	rnd         *rand.Rand
	training    []Sample
	validation  []Sample
	activations []ActivationFn
	weights     []Matrix
	biases      []Matrix
	wGradients  []Gradients
	bGradients  []Gradients
	threadData  []ThreadData
}

func NewTrainer(training, validation []Sample, topology []int, threads int, seed int64) *Trainer {
	var t = &Trainer{}
	t.topology = topology
	t.training = training
	t.validation = validation
	t.rnd = rand.New(rand.NewSource(seed))
	var layerSize = len(topology) - 1
	t.activations = make([]ActivationFn, layerSize)
	t.weights = make([]Matrix, layerSize)
	t.biases = make([]Matrix, layerSize)
	t.wGradients = make([]Gradients, layerSize)
	t.bGradients = make([]Gradients, layerSize)
	for layerIndex := 0; layerIndex < layerSize; layerIndex++ {
		if layerIndex == len(t.activations)-1 {
			t.activations[layerIndex] = &Sigmoid{SigmoidScale: SigmoidScale}
		} else {
			t.activations[layerIndex] = &ReLu{}
		}

		var inputSize = topology[layerIndex]
		var outputSize = topology[layerIndex+1]

		t.weights[layerIndex] = NewMatrix(outputSize, inputSize)
		t.biases[layerIndex] = NewMatrix(outputSize, 1)
		t.wGradients[layerIndex] = NewGradients(outputSize, inputSize)
		t.bGradients[layerIndex] = NewGradients(outputSize, 1)
	}
	t.threadData = make([]ThreadData, threads)
	for threadIndex := range t.threadData {
		var td = &t.threadData[threadIndex]
		td.wGradients = make([]Matrix, layerSize)
		td.bGradients = make([]Matrix, layerSize)
		td.neurons = make([][]Neuron, layerSize)
		for layerIndex := 0; layerIndex < layerSize; layerIndex++ {
			var inputSize = topology[layerIndex]
			var outputSize = topology[layerIndex+1]

			td.wGradients[layerIndex] = NewMatrix(outputSize, inputSize)
			td.bGradients[layerIndex] = NewMatrix(outputSize, 1)
			td.neurons[layerIndex] = make([]Neuron, outputSize)
		}
	}
	return t
}

func (t *Trainer) Train(epochs int, binFolderPath string) error {
	log.Println("Train started")
	defer log.Println("Train finished")

	t.initWeights()

	var bestValidationCost float64
	var bestEpoch int

	for epoch := 1; epoch <= epochs; epoch++ {
		t.startEpoch()
		log.Printf("Finished Epoch %v\n", epoch)

		validationCost := t.calcCost(t.validation)
		log.Printf("Current validation cost is: %f\n", validationCost)

		if bestEpoch == 0 ||
			validationCost < bestValidationCost {
			bestEpoch = epoch
			bestValidationCost = validationCost

			var err = t.saveNetwork(binFolderPath, epoch, validationCost)
			if err != nil {
				return err
			}
		} else {
			log.Printf("Best validation cost: %f Best epoch: %v\n", bestValidationCost, bestEpoch)
		}
	}

	return nil
}

func (t *Trainer) saveNetwork(binFolderPath string, epoch int, validationCost float64) error {
	var valCostInt = int(100000 * validationCost)
	var filepath = filepath.Join(binFolderPath, fmt.Sprintf("n-%2d-%v.nn", epoch, valCostInt))
	var err = t.makeNetwork().Save(filepath)
	if err != nil {
		return err
	}
	log.Println("Stored network", filepath)
	return nil
}

func (t *Trainer) calcCost(samples []Sample) float64 {
	var index int32 = -1
	var wg = &sync.WaitGroup{}
	var totalCost float64
	var mu = &sync.Mutex{}
	for i := range t.threadData {
		wg.Add(1)
		go func(td *ThreadData) {
			defer wg.Done()
			var localCost float64
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(samples) {
					break
				}
				sample := &samples[i]
				forward(t, td, sample)
				predicted := td.neurons[len(td.neurons)-1][0].A
				cost := ValidationCost(predicted, float64(sample.Target))
				localCost += cost
			}
			mu.Lock()
			totalCost += localCost
			mu.Unlock()
		}(&t.threadData[i])
	}
	wg.Wait()
	averageCost := totalCost / float64(len(samples))
	return averageCost
}

func (t *Trainer) initWeights() {
	for layerIndex := range t.activations {
		var inputSize = t.topology[layerIndex]
		var max = 1 / math.Sqrt(float64(inputSize))
		initUniform(t.rnd, t.weights[layerIndex].Data, max)
	}
}

func (t *Trainer) shuffle() {
	t.rnd.Shuffle(len(t.training), func(i, j int) {
		t.training[i], t.training[j] = t.training[j], t.training[i]
	})
}

func (t *Trainer) startEpoch() {
	t.shuffle()
	for i := 0; i+BatchSize <= len(t.training); i += BatchSize {
		t.trainBatch(t.training[i : i+BatchSize])
	}
}

func (t *Trainer) trainBatch(batch []Sample) {
	var index int32 = -1
	var wg = &sync.WaitGroup{}
	for i := range t.threadData {
		wg.Add(1)
		go func(td *ThreadData) {
			defer wg.Done()
			for layerIndex := range td.wGradients {
				td.wGradients[layerIndex].Reset()
				td.bGradients[layerIndex].Reset()
			}
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(batch) {
					break
				}
				sample := &batch[i]
				trainSample(t, td, sample)
			}
		}(&t.threadData[i])
	}
	wg.Wait()
	t.applyGradients()
}

func (t *Trainer) applyGradients() {
	for layerIndex := range t.weights {
		for threadIndex := range t.threadData {
			t.wGradients[layerIndex].AddMatrix(&t.threadData[threadIndex].wGradients[layerIndex])
			t.bGradients[layerIndex].AddMatrix(&t.threadData[threadIndex].bGradients[layerIndex])
		}
		t.wGradients[layerIndex].Apply(&t.weights[layerIndex])
		t.bGradients[layerIndex].Apply(&t.biases[layerIndex])
	}
}

func trainSample(t *Trainer, td *ThreadData, sample *Sample) {
	forward(t, td, sample)
	backward(t, td, sample)
}

func forward(t *Trainer, td *ThreadData, sample *Sample) {
	for layerIndex := range t.activations {
		var activation = t.activations[layerIndex]
		var weights = t.weights[layerIndex]
		var biases = t.biases[layerIndex]
		var neurons = td.neurons[layerIndex]
		if layerIndex == 0 {
			for outputIndex := range neurons {
				var x = 1 * biases.Data[outputIndex]
				for _, inputIndex := range sample.Input {
					x += 1 * weights.Get(outputIndex, int(inputIndex))
				}
				var n = &neurons[outputIndex]
				n.A = activation.Sigma(x)
				n.Prime = activation.SigmaPrime(x)
			}
		} else {
			var prevNeurons = td.neurons[layerIndex-1]
			for outputIndex := range neurons {
				var x = 1 * biases.Data[outputIndex]
				for inputIndex := range prevNeurons {
					x += prevNeurons[inputIndex].A * weights.Get(outputIndex, inputIndex)
				}
				var n = &neurons[outputIndex]
				n.A = activation.Sigma(x)
				n.Prime = activation.SigmaPrime(x)
			}
		}
	}
}

func backward(t *Trainer, td *ThreadData, sample *Sample) {
	// back propagation
	for layerIndex := len(t.activations) - 1; layerIndex >= 0; layerIndex-- {
		var neurons = td.neurons[layerIndex]
		var weights = t.weights[layerIndex]
		if layerIndex == len(t.activations)-1 {
			neurons[0].E = neurons[0].A - float64(sample.Target)
		} else {
			var nextNeurons = td.neurons[layerIndex+1]
			for i := range neurons {
				neurons[i].E = 0
			}
			for outputIndex := range nextNeurons {
				var n = &nextNeurons[outputIndex]
				var x = n.E * n.Prime
				for inputIndex := range neurons {
					neurons[inputIndex].E += weights.Get(outputIndex, inputIndex) * x
				}
			}
		}

		var wGradients = td.wGradients[layerIndex]
		var bGradients = td.bGradients[layerIndex]
		if layerIndex == 0 {
			for outputIndex := range neurons {
				var n = &neurons[outputIndex]
				var x = n.E * n.Prime
				bGradients.Data[outputIndex] += x * 1
				for _, inputIndex := range sample.Input {
					wGradients.Add(outputIndex, int(inputIndex), x*1)
				}
			}
		} else {
			var prevNeurons = td.neurons[layerIndex-1]
			for outputIndex := range neurons {
				var n = &neurons[outputIndex]
				var x = n.E * n.Prime
				bGradients.Data[outputIndex] += x * 1
				for inputIndex := range prevNeurons {
					wGradients.Add(outputIndex, inputIndex, x*prevNeurons[inputIndex].A)
				}
			}
		}
	}
}

func (t *Trainer) makeNetwork() *Network {
	var hiddenNeurons = make([]uint32, len(t.topology)-2)
	for i := range hiddenNeurons {
		hiddenNeurons[i] = uint32(t.topology[i+1])
	}
	var network = &Network{
		Id: 1,
		Topology: Topology{
			Inputs:        uint32(t.topology[0]),
			HiddenNeurons: hiddenNeurons,
			Outputs:       uint32(t.topology[len(t.topology)-1]),
		},
		Weights: t.weights,
		Biases:  t.biases,
	}
	return network
}
