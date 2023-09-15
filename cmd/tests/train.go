package main

import (
	"context"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/trainer"
)

func trainHandler() error {
	const sigmoidScale = 3.5 / 512
	var datasetProvider = &dataset.DatasetProvider{
		SigmoidScale:                sigmoidScale,
		MaxPosCount:                 8_000_000,
		GamesFolder:                 mapPath("~/chess/Dataset2023"),
		Threads:                     runtime.NumCPU(),
		SearchRatio:                 1.0,
		CheckNoisyOnlyForSideToMove: true,
	}
	return trainer.Run(context.Background(),
		datasetProvider, runtime.NumCPU(), 30, sigmoidScale, mapPath("~/chess/net"))
}
