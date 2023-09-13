package main

import (
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
)

const (
	BatchSize = 16384
)

type Trainer struct {
	weights    []float64
	gradients  []Gradient
	training   []Sample
	validation []Sample
	rnd        *rand.Rand
	threadData []ThreadData
}

type ThreadData struct {
	gradients []float64
}

func NewTrainer(training, validation []Sample, startingWeights []float64, threads int) *Trainer {
	var t = &Trainer{
		weights:    startingWeights,
		gradients:  make([]Gradient, len(startingWeights)),
		training:   training,
		validation: validation,
		rnd:        rand.New(rand.NewSource(0)),
		threadData: make([]ThreadData, threads),
	}
	for threadIndex := range t.threadData {
		var td = &t.threadData[threadIndex]
		td.gradients = make([]float64, len(startingWeights))
	}
	return t
}

func (t *Trainer) shuffle() {
	t.rnd.Shuffle(len(t.training), func(i, j int) {
		t.training[i], t.training[j] = t.training[j], t.training[i]
	})
}

func (t *Trainer) computeOutput(sample *Sample) float64 {
	var output, _ = t.computeOutput2(sample)
	return output
}

func (t *Trainer) computeOutput2(sample *Sample) (output, strongScale float64) {
	var mg = float64(sample.MiddleFree)
	var eg = float64(sample.EndgameFree)
	for _, f := range sample.Features {
		var val = float64(f.Value)
		mg += val * t.weights[2*f.Index]
		eg += val * t.weights[2*f.Index+1]
	}
	var mix = mg*float64(sample.MgPhase) + eg*float64(sample.EgPhase)
	if mix > 0 {
		strongScale = float64(sample.WhiteStrongScale)
	} else {
		strongScale = float64(sample.BlackStrongScale)
	}
	output = Sigmoid(strongScale * mix)
	return
}

func (t *Trainer) calcCost(samples []Sample) float64 {
	var totalCost float64
	var index int32 = -1
	var wg = &sync.WaitGroup{}
	var mu = &sync.Mutex{}
	for range t.threadData {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var localCost float64
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(samples) {
					break
				}
				sample := &samples[i]
				predicted := t.computeOutput(&samples[i])
				cost := ValidationCost(predicted, float64(sample.Target))
				localCost += cost
			}
			mu.Lock()
			totalCost += localCost
			mu.Unlock()
		}()
	}
	wg.Wait()
	averageCost := totalCost / float64(len(samples))
	return averageCost
}

func (t *Trainer) Train(epochs int) error {
	log.Println("Train started")
	defer log.Println("Train finished")

	for epoch := 1; epoch <= epochs; epoch++ {
		t.startEpoch()
		log.Printf("Finished Epoch %v\n", epoch)

		validationCost := t.calcCost(t.validation)
		log.Printf("Current validation cost is: %f\n", validationCost)

		/*if epoch%10 == 0 {
			trainingCost := t.calcCost(t.training)
			log.Printf("Current training cost is: %f\n", trainingCost)
		}*/
	}

	return nil
}

func (t *Trainer) startEpoch() {
	t.shuffle()
	for i := 0; i+BatchSize <= len(t.training); i += BatchSize {
		var batch = t.training[i : i+BatchSize]
		t.trainBatch(batch)
	}
}

func (t *Trainer) trainBatch(batch []Sample) {
	var index int32 = -1
	var wg = &sync.WaitGroup{}
	for i := range t.threadData {
		wg.Add(1)
		go func(td *ThreadData) {
			defer wg.Done()
			for i := range td.gradients {
				td.gradients[i] = 0
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

	for threadIndex := range t.threadData {
		var td = &t.threadData[threadIndex]
		for weightIndex := range t.gradients {
			t.gradients[weightIndex].Update(td.gradients[weightIndex])
		}
	}

	for weightIndex := range t.gradients {
		t.gradients[weightIndex].Apply(&t.weights[weightIndex])
	}
}

func trainSample(t *Trainer, td *ThreadData, sample *Sample) {
	var output, strongScale = t.computeOutput2(sample)
	outputGradient := strongScale * CalculateCostGradient(output, float64(sample.Target)) * SigmoidPrime(output)
	var mgOutputGradient = outputGradient * float64(sample.MgPhase)
	var egOutputGradient = outputGradient * float64(sample.EgPhase)
	for _, f := range sample.Features {
		var val = float64(f.Value)
		td.gradients[2*f.Index] += mgOutputGradient * val
		td.gradients[2*f.Index+1] += egOutputGradient * val
	}
}
