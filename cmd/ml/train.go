package main

import (
	"context"
	"log"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/trainer"
)

func trainHandler() error {
	var (
		gamesFolder    = mapPath("~/chess/Dataset2023")
		validationPath = mapPath("~/chess/tuner/quiet-labeled.epd")
		netFolderPath  = mapPath("~/chess/net")

		searchRatio                 = 1.0
		mirrorPos                   = true
		mergeRepeats                = true
		maxDatasetSize              = 1_000_000
		sigmoidScale                = 3.5 / 512
		inputSize, featureExtractor = 768, toFeatures768
		topology                    = []int{inputSize, 512, 1}
		epochs                      = 30
		concurrency                 = runtime.NumCPU()
	)

	var ctx = context.Background()
	var training, err = loadDatasetNN(ctx, featureExtractor, sigmoidScale,
		gamesFolder, searchRatio, mirrorPos, mergeRepeats, maxDatasetSize, concurrency)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset",
		"size", len(training))

	var validation []trainer.Sample
	if validationPath != "" {
		validation, err = loadValidationNN(validationPath, featureExtractor)
		if err != nil {
			return err
		}
		log.Println("Loaded validation dataset",
			"size", len(validation))
	}

	return trainer.Run(ctx, training, validation, topology,
		concurrency, epochs, sigmoidScale, netFolderPath)
}
