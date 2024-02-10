package main

import (
	"context"
	"log"
	"sort"
	"sync"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/internal/pgn"
	"github.com/ChizhovVadim/CounterGo/internal/trainer"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

func loadDatasetNN(
	ctx context.Context,
	featureExtractor func(pos *common.Position) []domain.FeatureInfo,
	sigmoidScale float64,
	gamesFolder string,
	searchRatio float64,
	mirrorPos bool,
	mergeRepeats bool,
	maxSize int,
	concurrency int,
) ([]trainer.Sample, error) {

	var result []trainer.Sample
	var games = make(chan pgn.GameRaw, 16)
	var samples = make(chan []trainer.Sample, 128)
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
			return extractFeaturesNN(ctx, sigmoidScale, searchRatio, featureExtractor, mirrorPos, games, samples)
		})
	}
	g.Go(func() error {
		wg.Wait()
		close(samples)
		return nil
	})
	g.Go(func() error {
		var err error
		result, err = collectSamplesNN(ctx, samples, mergeRepeats, maxSize, datasetReady)
		return err
	})
	return result, g.Wait()
}

func collectSamplesNN(
	ctx context.Context,
	samples <-chan []trainer.Sample,
	mergeRepeats bool,
	maxSize int,
	datasetReady chan<- struct{},
) ([]trainer.Sample, error) {
	var result []trainer.Sample

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
		result = joinEqualFeatures(result)
	}

	return result, nil
}

func extractFeaturesNN(
	ctx context.Context,
	sigmoidScale float64,
	searchRatio float64,
	featureExtractor func(pos *common.Position) []domain.FeatureInfo,
	mirrorPos bool,
	games <-chan pgn.GameRaw,
	samples chan<- []trainer.Sample,
) error {
	for gameRaw := range games {
		var chunk []trainer.Sample
		var err = dataset.AnalyzeGame(sigmoidScale, searchRatio, gameRaw, func(di dataset.DatasetItem2) error {
			var features = featureExtractor(di.Pos)
			sortFeatures(features)
			chunk = append(chunk, trainer.Sample{
				Input:  features,
				Target: float32(di.Target),
			})

			if mirrorPos {
				var mirror = common.MirrorPosition(di.Pos)
				var features = featureExtractor(&mirror)
				sortFeatures(features)
				chunk = append(chunk, trainer.Sample{
					Input:  features,
					Target: float32(1 - di.Target),
				})
			}

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

func joinEqualFeatures(samples []trainer.Sample) []trainer.Sample {
	sort.Slice(samples, func(i, j int) bool {
		return compareFeatures(samples[i].Input, samples[j].Input) < 0
	})

	var result []trainer.Sample
	var sum float64
	var count int
	var repeatCount int

	for i := range samples {
		var item = &samples[i]
		if i > 0 && compareSamplesNN(item, &samples[i-1]) == 0 {
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

	log.Println("repeat count", repeatCount)

	return result
}

func compareSamplesNN(l, r *trainer.Sample) int {
	return compareFeatures(l.Input, r.Input)
}

func loadValidationNN(
	filepath string,
	f func(pos *common.Position) []domain.FeatureInfo,
) ([]trainer.Sample, error) {
	var data, err = dataset.LoadZurichessDataset(filepath)
	if err != nil {
		return nil, err
	}

	var result []trainer.Sample

	for _, item := range data {
		var pos, err = common.NewPositionFromFEN(item.Fen)
		if err != nil {
			return nil, err
		}
		var features = f(&pos)
		result = append(result, trainer.Sample{Input: features, Target: float32(item.Target)})
	}

	return result, nil
}
