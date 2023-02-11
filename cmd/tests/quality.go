package main

import (
	"log"
	"math"

	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func runCheckEvalQuality(evalName string, datasetPath string) error {
	logger.Println("checkEvalQuality started",
		"evalName", evalName,
		"datasetPath", datasetPath)
	defer logger.Println("checkEvalQuality finished")

	var evaluator = evalbuilder.Get(evalName)().(Evaluator)
	return checkEvalQuality(evaluator, datasetPath)
}

func checkEvalQuality(e Evaluator, datasetPath string) error {
	var totalCost float64
	var count int
	var checkSymmetricEval = true

	var err = walkDataset(datasetPath, func(pos *common.Position, gameResult float64) error {
		var score = e.Evaluate(pos)

		var whitePointOfViewScore = score
		if !pos.WhiteMove {
			whitePointOfViewScore = -whitePointOfViewScore
		}
		var x = Sigmoid(float64(whitePointOfViewScore)) - gameResult
		totalCost += x * x
		count++

		if checkSymmetricEval {
			var mirrorPos = common.MirrorPosition(pos)
			if e.Evaluate(&mirrorPos) != score {
				checkSymmetricEval = false
				log.Println("Eval not symmetric", pos.String())
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	var averageCost = totalCost / float64(count)
	log.Printf("Average cost: %f", averageCost)
	return nil
}

const SigmoidScale = 3.5 / 512

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(SigmoidScale*(-x)))
}
