package main

import (
	"log"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/internal/tuner"
)

func tunerHandler(args []string) error {
	var (
		evalName        = "counter"
		gamesFolderPath = mapPath("~/chess/Dataset2023")
		sigmoidScale    = 0.011
		searchRatio     = 1.0
		maxDatasetSize  = 6_000_000
		epochs          = 15
		concurrency     = runtime.NumCPU()
	)
	var buildEvalService = func() tuner.IFeatureProvider {
		return evalbuilder.Get(evalName)().(tuner.IFeatureProvider)
	}
	samples, err := tuner.LoadDataset(buildEvalService,
		gamesFolderPath, sigmoidScale, searchRatio, maxDatasetSize, concurrency)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset",
		"size", len(samples))
	return tuner.RunTuner(buildEvalService().FeatureSize(), samples, epochs, concurrency)
}
