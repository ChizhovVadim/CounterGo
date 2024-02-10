package main

import (
	"log"
	"math"

	"github.com/ChizhovVadim/CounterGo/internal/dataset"
	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type IEvaluator interface {
	Evaluate(p *common.Position) int
}

func qualityHandler() error {
	var (
		evalName       = cliArgs.GetString("eval", "")
		sigmoidScale   = 3.5 / 512
		validationPath = mapPath("~/chess/tuner/quiet-labeled.epd")
	)

	log.Println("checkEvalQuality started",
		"evalName", evalName)
	defer log.Println("checkEvalQuality finished")

	var evaluator = evalbuilder.Get(evalName)().(IEvaluator)
	var data, err = dataset.LoadZurichessDataset(validationPath)
	if err != nil {
		return err
	}
	return checkEvalQuality(evaluator, sigmoidScale, data)
}

func checkEvalQuality(
	e IEvaluator,
	sigmoidScale float64,
	dataset []dataset.DatasetItem,
) error {
	var totalCost float64
	var count int
	var checkSymmetricEval = true

	for _, item := range dataset {
		var pos, err = common.NewPositionFromFEN(item.Fen)
		if err != nil {
			return err
		}
		var score = e.Evaluate(&pos)
		var whitePointOfViewScore = score
		if !pos.WhiteMove {
			whitePointOfViewScore = -whitePointOfViewScore
		}
		var x = sigmoid(float64(whitePointOfViewScore), sigmoidScale) - item.Target
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
}

func sigmoid(x, sigmoidScale float64) float64 {
	return 1.0 / (1.0 + math.Exp(sigmoidScale*(-x)))
}
