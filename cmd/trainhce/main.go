package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
	eval "github.com/ChizhovVadim/CounterGo/pkg/eval/linear"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type ITunableEvaluator interface {
	StartingWeights() []float64
	ComputeFeatures(pos *common.Position) domain.TuneEntry
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var e = eval.NewEvaluationService()
	var trainingPath = "/Users/vadimchizhov/chess/fengen.txt"
	var validationPath = "/Users/vadimchizhov/chess/tuner/quiet-labeled.epd"
	var threads = 4
	var epochs = 100

	var err = run(e, trainingPath, validationPath, threads, epochs)
	if err != nil {
		log.Println(err)
	}
}

func run(evaluator ITunableEvaluator, trainingPath, validationPath string, threads, epochs int) error {
	td, err := LoadDataset(trainingPath, evaluator, parseTrainingSample)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset", len(td))
	runtime.GC()

	vd, err := LoadDataset(validationPath, evaluator, parseValidationSample)
	if err != nil {
		return err
	}
	log.Println("Loaded validation", len(vd))

	var weights = evaluator.StartingWeights()
	log.Println("Num of weights", len(weights))

	var trainer = &Trainer{
		threads:    threads,
		weigths:    weights,
		gradients:  make([]Gradient, len(weights)),
		training:   td,
		validation: vd,
		rnd:        rand.New(rand.NewSource(0)),
	}

	err = trainer.Train(epochs)
	if err != nil {
		return err
	}

	var wInt = make([]int, len(trainer.weigths))
	for i := range wInt {
		wInt[i] = int(math.Round(100 * trainer.weigths[i]))
	}
	fmt.Printf("var w = %#v\n", wInt)

	return nil
}
