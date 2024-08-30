package main

import (
	"log"
	"math/rand"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/ml"
	"github.com/ChizhovVadim/CounterGo/internal/train"
)

func trainHandler(args []string) error {
	var (
		gamesFolderPath = mapPath("~/chess/Dataset2023")
		netFolderPath   = mapPath("~/chess/net")
		sigmoidScale    = 0.011
		searchRatio     = 1.0
		maxDatasetSize  = 50_000_000
		epochs          = 15
		concurrency     = runtime.NumCPU()
		mirrorPos       = true
		costName        = "mse"
		hiddenSize      = 512
	)
	var buildFeatureService = func() train.IFeatureProvider {
		return &train.Feature768Provider{}
	}
	samples, err := train.LoadDataset(buildFeatureService,
		gamesFolderPath, sigmoidScale, searchRatio, maxDatasetSize, concurrency, mirrorPos)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset",
		"size", len(samples))
	var model = train.NewModel(buildFeatureService().FeatureSize(), hiddenSize)
	model.InitWeights(rand.New(rand.NewSource(0)))
	cost, err := ml.NewCost(costName)
	if err != nil {
		return err
	}
	return train.Train(samples, epochs, model, cost, concurrency, netFolderPath)
}
