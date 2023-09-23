package trainer

import (
	"context"
	"log"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
)

type IDatasetProvider interface {
	Load(ctx context.Context, dataset chan<- domain.DatasetItem) error
}

func Run(
	ctx context.Context,
	datasetProvider IDatasetProvider,
	validationProvider IDatasetProvider,
	threads int,
	epochs int,
	sigmoidScale float64,
	netFolderPath string,
) error {
	dataset, err := loadDataset(ctx, datasetProvider, true)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset", len(dataset))
	runtime.GC()

	var training, validation []Sample
	if validationProvider != nil {
		validation, err = loadDataset(ctx, validationProvider, false)
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

	var trainer = NewTrainer(training, validation, []int{769, 512, 1}, threads, 0, sigmoidScale)
	return trainer.Train(epochs, netFolderPath)
}
