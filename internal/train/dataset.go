package train

import (
	"context"
	"sync"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/math"
	"github.com/ChizhovVadim/CounterGo/internal/pgn"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

func LoadDataset(
	featureProvider func() IFeatureProvider,
	gamesFolder string,
	sigmoidScale float64,
	searchRatio float64,
	maxSize int,
	concurrency int,
	mirrorPos bool,
) ([]Sample, error) {
	var datasetReady = make(chan struct{})
	var games = make(chan pgn.GameRaw, 16)
	var results = make(chan []Sample, 16)

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		defer close(games)
		return dataset.LoadGames(ctx, gamesFolder, datasetReady, games)
	})

	var res []Sample
	g.Go(func() error {
		var samples, err = collectResult(ctx, results, maxSize, datasetReady)
		res = samples
		return err
	})

	var wg = &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return analyzeGames(ctx, games, results, featureProvider(), sigmoidScale, searchRatio, mirrorPos)
		})
	}

	g.Go(func() error {
		wg.Wait()
		close(results)
		return nil
	})

	var err = g.Wait()
	return res, err
}

func collectResult(
	ctx context.Context,
	samples <-chan []Sample,
	maxSize int,
	datasetReady chan<- struct{},
) ([]Sample, error) {
	var res []Sample
	for chunk := range samples {
		res = append(res, chunk...)
		if len(res) >= maxSize && datasetReady != nil {
			close(datasetReady)
			datasetReady = nil
		}
	}
	return res, nil
}

func analyzeGames(
	ctx context.Context,
	games <-chan pgn.GameRaw,
	samples chan<- []Sample,
	featureProvider IFeatureProvider,
	sigmoidScale float64,
	searchRatio float64,
	mirrorPos bool,
) error {
	for gameRaw := range games {
		game, err := dataset.AnalyzeGame(gameRaw)
		if err != nil {
			return err
		}
		var chunk []Sample
		for i := range game.Positions {
			var pos = &game.Positions[i]
			var features = featureProvider.ComputeFeatures(&pos.Position)
			var target = computeTarget(pos.Position.WhiteMove, pos.ScoreMate, pos.ScoreCentipawns, sigmoidScale, searchRatio, game.GameResult)
			chunk = append(chunk, Sample{
				Target:    float32(target),
				TuneEntry: features,
			})
			if mirrorPos {
				var mirror = common.MirrorPosition(&pos.Position)
				var mirrorFeatures = featureProvider.ComputeFeatures(&mirror)
				var mirrorTarget = 1 - target
				chunk = append(chunk, Sample{
					Target:    float32(mirrorTarget),
					TuneEntry: mirrorFeatures,
				})
			}
		}
		if len(chunk) != 0 {
			samples <- chunk
		}
	}
	return nil
}

func computeTarget(
	wstm bool,
	scoreMate int,
	scoreCentipawns int,
	sigmoidScale float64,
	searchRatio float64,
	gameResult float64,
) float64 {
	var targetBySearch float64
	if scoreMate != 0 {
		if scoreMate > 0 {
			targetBySearch = 1
		} else {
			targetBySearch = 0
		}
	} else {
		targetBySearch = math.Sigmoid(sigmoidScale * float64(scoreCentipawns))
	}
	if !wstm {
		targetBySearch = 1 - targetBySearch
	}
	return targetBySearch*searchRatio + gameResult*(1-searchRatio)
}
