package main

import (
	"context"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/tuner"
	eval "github.com/ChizhovVadim/CounterGo/pkg/eval/fast"
)

func tunerHandler() error {
	const sigmoidScale = 3.5 / 512
	var datasetProvider = &dataset.DatasetProvider{
		SigmoidScale:                sigmoidScale,
		MaxPosCount:                 8_000_000,
		GamesFolder:                 mapPath("~/chess/Dataset2023"),
		Threads:                     runtime.NumCPU(),
		SearchRatio:                 1.0,
		CheckNoisyOnlyForSideToMove: true,
	}
	var evaluator = eval.NewEvaluationService()
	return tuner.Run(context.Background(), datasetProvider, evaluator, runtime.NumCPU(), 100, sigmoidScale)
}
