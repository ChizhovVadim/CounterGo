package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/internal/evalbuilder"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type ITunableEvaluator interface {
	EnableTuning()
	StartingWeights() []float64
	ComputeFeatures(pos *common.Position) domain.TuneEntry
}

type Config struct {
	eval           string
	trainingPath   string
	validationPath string
	threads        int
	epochs         int
	searchWeight   float64
	datasetMaxSize int
}

var config Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&config.eval, "eval", "", "Eval function to train")
	flag.StringVar(&config.trainingPath, "td", "", "Path to training dataset")
	flag.StringVar(&config.validationPath, "vd", "", "Path to validation dataset")
	flag.IntVar(&config.threads, "threads", 1, "Number of threads")
	flag.IntVar(&config.epochs, "epochs", 5, "Number of epochs")
	flag.Float64Var(&config.searchWeight, "sw", 0.5, "Weight of search result in training dataset")
	flag.IntVar(&config.datasetMaxSize, "dms", 1000000, "Max size of dataset")
	flag.Parse()

	log.Printf("%+v", config)

	var e = evalbuilder.NewEvalBuilder(config.eval)().(ITunableEvaluator)
	e.EnableTuning()

	var err = run(e)
	if err != nil {
		log.Println(err)
	}
}

func run(evaluator ITunableEvaluator) error {
	td, err := LoadDataset(config.trainingPath, evaluator, parseTrainingSample)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset", len(td))
	runtime.GC()

	vd, err := LoadDataset(config.validationPath, evaluator, parseValidationSample)
	if err != nil {
		return err
	}
	log.Println("Loaded validation", len(vd))

	var weights = evaluator.StartingWeights()
	log.Println("Num of weights", len(weights))

	var trainer = &Trainer{
		threads:    config.threads,
		weigths:    weights,
		gradients:  make([]Gradient, len(weights)),
		training:   td,
		validation: vd,
		rnd:        rand.New(rand.NewSource(0)),
	}

	err = trainer.Train(config.epochs)
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
