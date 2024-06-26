package main

import (
	"context"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/internal/tuner"
)

func tunerHandler() error {
	const evalName = "counter"
	const sigmoidScale = 3.5 / 512
	var datasetProvider = &dataset.DatasetProvider{
		SigmoidScale: sigmoidScale,
		MaxPosCount:  8_000_000,
		GamesFolder:  mapPath("~/chess/Dataset2023"),
		Threads:      runtime.NumCPU(),
		SearchRatio:  1.0,
	}
	var evaluator = evalbuilder.Get(evalName)().(tuner.ITunableEvaluator)
	return tuner.Run(context.Background(), datasetProvider,
		newValidationDatasetProvider(),
		evaluator, runtime.NumCPU(), 100, sigmoidScale)
}
