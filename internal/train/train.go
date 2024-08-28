package train

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

func Train(
	samples []Sample,
	epochs int,
	mainModel *Model,
	concurrency int,
	netFolderPath string,
) error {
	log.Println("Train started")
	defer log.Println("Train finished")

	err := os.MkdirAll(netFolderPath, os.ModePerm)
	if err != nil {
		return err
	}

	var validationSize = min(500_000, len(samples)/5)
	var validation = samples[:validationSize]
	var training = samples[validationSize:]

	const BatchSize = 16384
	var models = make([]*Model, concurrency)
	models[0] = mainModel
	for i := 1; i < len(models); i++ {
		models[i] = mainModel.ThreadCopy()
	}

	var rnd = rand.New(rand.NewSource(0))
	for epoch := 1; epoch <= epochs; epoch++ {
		shuffle(rnd, training)
		for i := 0; i+BatchSize <= len(training); i += BatchSize {
			var batch = training[i : i+BatchSize]
			trainBatch(batch, models)
			applyGradients(models)
		}
		log.Printf("Finished Epoch %v\n", epoch)
		validationCost := calcAverageCost(validation, models)
		log.Printf("Current validation cost is: %f\n", validationCost)
		if netFolderPath != "" {
			var err = models[0].Save(buildNetPath(netFolderPath, epoch, validationCost))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func shuffle(rnd *rand.Rand, training []Sample) {
	rnd.Shuffle(len(training), func(i, j int) {
		training[i], training[j] = training[j], training[i]
	})
}

func trainBatch(samples []Sample, models []*Model) {
	var index int32 = -1
	var wg = &sync.WaitGroup{}
	for i := range models {
		wg.Add(1)
		go func(m *Model) {
			defer wg.Done()
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(samples) {
					break
				}
				sample := &samples[i]
				m.Train(sample)
			}
		}(models[i])
	}
	wg.Wait()
}

func applyGradients(models []*Model) {
	for i := 1; i < len(models); i++ {
		models[i].AddGradients(models[0])
	}
	models[0].ApplyGradients()
}

func calcAverageCost(samples []Sample, models []*Model) float64 {
	var index int32 = -1
	var wg = &sync.WaitGroup{}
	var totalCost float64
	var mu = &sync.Mutex{}
	for i := range models {
		wg.Add(1)
		go func(m *Model) {
			defer wg.Done()
			var localCost float64
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(samples) {
					break
				}
				localCost += m.CalcCost(&samples[i])
			}
			mu.Lock()
			totalCost += localCost
			mu.Unlock()
		}(models[i])
	}
	wg.Wait()
	averageCost := totalCost / float64(len(samples))
	return averageCost
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buildNetPath(netFolderPath string, epoch int, validationCost float64) string {
	var valCostInt = int(100000 * validationCost)
	//TODO insert date
	return filepath.Join(netFolderPath, fmt.Sprintf("n-%2d-%v.nn", epoch, valCostInt))
}
