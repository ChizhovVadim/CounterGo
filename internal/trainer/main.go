package trainer

import (
	"context"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
)

type Sample struct {
	Input  []domain.FeatureInfo
	Target float32
}

func Run(
	ctx context.Context,
	training, validation []Sample,
	topology []int,
	threads int,
	epochs int,
	sigmoidScale float64,
	netFolderPath string,
) error {
	if len(validation) == 0 {
		var validationSize = min(500_000, len(training)/5)
		validation = training[:validationSize]
		training = training[validationSize:]
	}
	var trainer = NewTrainer(training, validation, topology, threads, 0, sigmoidScale)
	return trainer.Train(epochs, netFolderPath)
}
