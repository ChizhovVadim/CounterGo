package main

import (
	"context"
	"log"
	"runtime"
	"sort"
	"sync"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/internal/pgn"
	"github.com/ChizhovVadim/CounterGo/internal/tuner"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	eval "github.com/ChizhovVadim/CounterGo/pkg/eval/counter"
	"golang.org/x/sync/errgroup"
)

type ITunableEvaluator interface {
	ComputeFeatures(pos *common.Position) domain.TuneEntry
	StartingWeights() []float64
}

func newEvalService() ITunableEvaluator {
	var evalService = eval.NewEvaluationService()
	evalService.EnableTuning()
	return evalService
}

func tunerHandler() error {
	var (
		gamesFolder    = mapPath("~/chess/Dataset2023")
		validationPath = mapPath("~/chess/tuner/quiet-labeled.epd")
		searchRatio    = 1.0
		mergeRepeats   = true
		epochs         = 100
		sigmoidScale   = 3.5 / 512
		maxDatasetSize = 1_000_000
		concurrency    = runtime.NumCPU()
	)

	var evalService = newEvalService()
	var startingWeights = evalService.StartingWeights()

	var ctx = context.Background()
	var training, err = loadDatasetHCE(ctx, sigmoidScale, gamesFolder, searchRatio, mergeRepeats, maxDatasetSize, concurrency)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset",
		"size", len(training))
	var validation []tuner.Sample
	if validationPath != "" {
		validation, err = loadValidationHCE(validationPath, evalService)
		if err != nil {
			return err
		}
		log.Println("Loaded validation dataset",
			"size", len(validation))
	}

	return tuner.Run(ctx, training, validation, concurrency, epochs, sigmoidScale, startingWeights)
}

func loadDatasetHCE(
	ctx context.Context,
	sigmoidScale float64,
	gamesFolder string,
	searchRatio float64,
	mergeRepeats bool,
	maxSize int,
	concurrency int,
) ([]tuner.Sample, error) {
	var result []tuner.Sample
	var games = make(chan pgn.GameRaw, 16)
	var samples = make(chan []tuner.Sample, 128)
	var datasetReady = make(chan struct{})

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(games)
		return dataset.LoadGames(ctx, gamesFolder, games, datasetReady)
	})
	var wg = &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			var evaluator = newEvalService()
			return extractFeaturesHCE(ctx, sigmoidScale, searchRatio, evaluator, games, samples)
		})
	}
	g.Go(func() error {
		wg.Wait()
		close(samples)
		return nil
	})
	g.Go(func() error {
		var err error
		result, err = collectSamplesHCE(ctx, samples, mergeRepeats, maxSize, datasetReady)
		return err
	})
	return result, g.Wait()
}

func collectSamplesHCE(
	ctx context.Context,
	samples <-chan []tuner.Sample,
	mergeRepeats bool,
	maxSize int,
	datasetReady chan<- struct{},
) ([]tuner.Sample, error) {
	var result []tuner.Sample

	for sample := range samples {
		result = append(result, sample...)
		if maxSize != 0 && len(result) >= maxSize {
			if datasetReady != nil {
				log.Println("skip rest positions")
				close(datasetReady)
				datasetReady = nil
			}
		}
	}

	if mergeRepeats {
		log.Println("Merge repeats")
		result = joinEqualFeaturesHCE(result)
	}

	return result, nil
}

func extractFeaturesHCE(
	ctx context.Context,
	sigmoidScale float64,
	searchRatio float64,
	evaluator ITunableEvaluator,
	games <-chan pgn.GameRaw,
	samples chan<- []tuner.Sample,
) error {
	for gameRaw := range games {
		var chunk []tuner.Sample
		var err = dataset.AnalyzeGame(sigmoidScale, searchRatio, gameRaw, func(di dataset.DatasetItem2) error {
			var featureSet = evaluator.ComputeFeatures(di.Pos)
			sortFeatures(featureSet.Features)
			var sample = tuner.Sample{
				Target:    float32(di.Target),
				TuneEntry: featureSet,
			}
			chunk = append(chunk, sample)
			return nil
		})
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case samples <- chunk:
		}
	}
	return nil
}

// Так как кол-во весов в 2 раза больше кол-ва признаков,
// то одинаковые признаки, только когда phase совпадают.
func compareSamplesHCE(l, r *tuner.Sample) int {
	if l.MgPhase < r.MgPhase {
		return -1
	}
	if l.MgPhase > r.MgPhase {
		return 1
	}
	return compareFeatures(l.Features, r.Features)
}

func joinEqualFeaturesHCE(samples []tuner.Sample) []tuner.Sample {
	sort.Slice(samples, func(i, j int) bool {
		return compareSamplesHCE(&samples[i], &samples[j]) < 0
	})

	var result []tuner.Sample
	var sum float64
	var count int
	var repeatCount int

	for i := range samples {
		var item = &samples[i]
		if i > 0 && compareSamplesHCE(item, &samples[i-1]) == 0 {
			// same features
			sum += float64(item.Target)
			count += 1
			repeatCount += 1
		} else {
			// new features
			if i > 0 {
				var copy = samples[i-1]
				copy.Target = float32(sum / float64(count))
				result = append(result, copy)
			}
			sum = float64(item.Target)
			count = 1
		}
	}

	if count > 0 {
		var copy = samples[len(samples)-1]
		copy.Target = float32(sum / float64(count))
		result = append(result, copy)
	}

	log.Println("repeat count", repeatCount)

	return result
}

func loadValidationHCE(
	filepath string,
	evaluator ITunableEvaluator,
) ([]tuner.Sample, error) {
	var data, err = dataset.LoadZurichessDataset(filepath)
	if err != nil {
		return nil, err
	}

	var result []tuner.Sample

	for _, item := range data {
		var pos, err = common.NewPositionFromFEN(item.Fen)
		if err != nil {
			return nil, err
		}
		var featureSet = evaluator.ComputeFeatures(&pos)
		result = append(result, tuner.Sample{
			Target:    float32(item.Target),
			TuneEntry: featureSet,
		})
	}

	return result, nil
}
