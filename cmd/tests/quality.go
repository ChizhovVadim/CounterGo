package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

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
	file, err := os.Open(datasetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var totalCost float64
	var count int
	var checkSymmetricEval = true

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()

		var index = strings.Index(s, "\"")
		if index < 0 {
			return fmt.Errorf("zurichessParser failed %v", s)
		}

		var fen = s[:index]
		pos, err := common.NewPositionFromFEN(fen)
		if err != nil {
			return err
		}

		var strScore = s[index+1:]
		var prob float64
		if strings.HasPrefix(strScore, "1/2-1/2") {
			prob = 0.5
		} else if strings.HasPrefix(strScore, "1-0") {
			prob = 1.0
		} else if strings.HasPrefix(strScore, "0-1") {
			prob = 0.0
		} else {
			return fmt.Errorf("zurichessParser failed %v", s)
		}

		var score = e.Evaluate(&pos)

		var whitePointOfViewScore = score
		if !pos.WhiteMove {
			whitePointOfViewScore = -whitePointOfViewScore
		}
		var x = Sigmoid(float64(whitePointOfViewScore)) - prob
		totalCost += x * x
		count++

		if checkSymmetricEval {
			var mirrorPos = common.MirrorPosition(&pos)
			if e.Evaluate(&mirrorPos) != score {
				checkSymmetricEval = false
				log.Println("Eval not symmetric", pos.String())
			}
		}
	}
	var averageCost = totalCost / float64(count)
	log.Printf("Average cost: %f", averageCost)
	return nil
}

const SigmoidScale = 3.5 / 512

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(SigmoidScale*(-x)))
}
