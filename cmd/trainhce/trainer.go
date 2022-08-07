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
	threads    int
	weigths    []float64
	gradients  []Gradient
	training   []Sample
	validation []Sample
	rnd        *rand.Rand
}

func (t *Trainer) Shuffle() {
	t.rnd.Shuffle(len(t.training), func(i, j int) {
		t.training[i], t.training[j] = t.training[j], t.training[i]
	})
}

func (t *Trainer) computeOutput(sample *Sample) float64 {
	var output, _ = t.computeOutput2(sample)
	return output
}

func (t *Trainer) computeOutput2(sample *Sample) (output, strongScale float64) {
	var mg, eg float64
	for _, f := range sample.Features {
		var val = float64(f.Value)
		mg += val * t.weigths[2*f.Index]
		eg += val * t.weigths[2*f.Index+1]
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
	for i := 0; i < t.threads; i++ {
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
		t.StartEpoch()
		log.Printf("Finished Epoch %v\n", epoch)

		validationCost := t.calcCost(t.validation)
		log.Printf("Current validation cost is: %f\n", validationCost)

		if epoch%10 == 0 {
			trainingCost := t.calcCost(t.training)
			log.Printf("Current training cost is: %f\n", trainingCost)
		}
	}

	return nil
}

func (t *Trainer) StartEpoch() {
	t.Shuffle()
	for i := 0; i+BatchSize <= len(t.training); i += BatchSize {
		var batch = t.training[i : i+BatchSize]
		t.trainBatch(batch)
	}
}

func (t *Trainer) trainBatch(samples []Sample) {
	for weightIndex := range t.gradients {
		t.gradients[weightIndex].Reset()
	}

	for i := range samples {
		var sample = &samples[i]
		var output, strongScale = t.computeOutput2(sample)
		outputGradient := strongScale * CalculateCostGradient(output, float64(sample.Target)) * SigmoidPrime(output)
		var mgOutputGradient = outputGradient * float64(sample.MgPhase)
		var egOutputGradient = outputGradient * float64(sample.EgPhase)
		for _, f := range sample.Features {
			var val = float64(f.Value)
			t.gradients[2*f.Index].Update(mgOutputGradient * val)
			t.gradients[2*f.Index+1].Update(egOutputGradient * val)
		}
	}

	for weightIndex := range t.gradients {
		t.gradients[weightIndex].Apply(&t.weigths[weightIndex])
	}
}
