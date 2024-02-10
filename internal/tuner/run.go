package tuner

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
)

type Sample struct {
	domain.TuneEntry
	Target float32
}

func Run(
	ctx context.Context,
	training, validation []Sample,
	threads int,
	epochs int,
	sigmoidScale float64,
	startingWeights []float64,
) error {
	if len(validation) == 0 {
		var validationSize = min(500_000, len(training)/5)
		validation = training[:validationSize]
		training = training[validationSize:]
	}

	var weights = startingWeights
	log.Println("Num of weights", len(weights))

	var trainer = NewTrainer(training, validation, weights, threads, sigmoidScale)
	var err = trainer.Train(epochs)
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
