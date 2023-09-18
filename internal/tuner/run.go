package tuner

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type IDatasetProvider interface {
	Load(ctx context.Context, dataset chan<- domain.DatasetItem) error
}

type ITunableEvaluator interface {
	EnableTuning()
	StartingWeights() []float64
	ComputeFeatures(pos *common.Position) domain.TuneEntry
}

func Run(
	ctx context.Context,
	datasetProvider IDatasetProvider,
	validationProvider IDatasetProvider,
	tunableEvaluator ITunableEvaluator,
	threads int,
	epochs int,
	sigmoidScale float64,
) error {

	dataset, err := loadDataset(ctx, datasetProvider, tunableEvaluator)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset", len(dataset))
	runtime.GC()

	var training, validation []Sample
	if validationProvider != nil {
		validation, err = loadDataset(ctx, validationProvider, tunableEvaluator)
		if err != nil {
			return err
		}
		log.Println("Loaded validation dataset", len(validation))
		training = dataset
	} else {
		var validationSize = min(500_000, len(dataset)/5)
		validation = dataset[:validationSize]
		training = dataset[validationSize:]
	}

	var weights = tunableEvaluator.StartingWeights()
	log.Println("Num of weights", len(weights))

	var trainer = NewTrainer(training, validation, weights, threads, sigmoidScale)
	err = trainer.Train(epochs)
	if err != nil {
		return err
	}

	var wInt = make([]int, len(trainer.weights))
	for i := range wInt {
		wInt[i] = int(math.Round(100 * trainer.weights[i]))
	}
	fmt.Printf("var w = %#v\n", wInt)

	return nil
}
