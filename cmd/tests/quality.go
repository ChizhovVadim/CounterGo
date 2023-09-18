package main

import (
	"context"
	"log"
	"math"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

type IDatasetProvider interface {
	Load(ctx context.Context, dataset chan<- domain.DatasetItem) error
}

func runCheckEvalQuality(evalName string, datasetPath string) error {
	logger.Println("checkEvalQuality started",
		"evalName", evalName,
		"datasetPath", datasetPath)
	defer logger.Println("checkEvalQuality finished")

	var evaluator = evalbuilder.Get(evalName)().(Evaluator)
	var datasetProvider = &dataset.ZurichessDatasetProvider{
		FilePath: datasetPath,
	}
	return checkEvalQuality(context.Background(), evaluator, datasetProvider)
}

func checkEvalQuality(
	ctx context.Context,
	e Evaluator,
	datasetProvider IDatasetProvider,
) error {
	g, ctx := errgroup.WithContext(ctx)
	var dataset = make(chan domain.DatasetItem, 128)
	g.Go(func() error {
		defer close(dataset)
		return datasetProvider.Load(ctx, dataset)
	})
	g.Go(func() error {
		var totalCost float64
		var count int
		var checkSymmetricEval = true

		for item := range dataset {
			var pos, err = common.NewPositionFromFEN(item.Fen)
			if err != nil {
				return err
			}
			var score = e.Evaluate(&pos)
			var whitePointOfViewScore = score
			if !pos.WhiteMove {
				whitePointOfViewScore = -whitePointOfViewScore
			}
			var x = Sigmoid(float64(whitePointOfViewScore)) - item.Target
			totalCost += x * x
			count++

			if checkSymmetricEval {
				var mirrorPos = common.MirrorPosition(&pos)
				if e.Evaluate(&mirrorPos) != score {
					checkSymmetricEval = false
					log.Println("Eval not symmetric", item.Fen)
				}
			}
		}

		var averageCost = totalCost / float64(count)
		log.Printf("Average cost: %f", averageCost)
		return nil
	})
	return g.Wait()
}

func Sigmoid(x float64) float64 {
	const SigmoidScale = 3.5 / 512
	return 1.0 / (1.0 + math.Exp(SigmoidScale*(-x)))
}
