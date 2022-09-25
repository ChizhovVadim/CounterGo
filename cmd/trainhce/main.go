package main

import (
	"flag"
	"fmt"
	"log"
	"math"
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

	var e = evalbuilder.Build(config.eval).(ITunableEvaluator)
	e.EnableTuning()

	var err = run(e)
	if err != nil {
		log.Println(err)
	}
}

func run(evaluator ITunableEvaluator) error {
	dataset, err := LoadDataset(config.trainingPath, evaluator, parseTrainingSample)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset", len(dataset))
	runtime.GC()

	var training, validation []Sample
	if config.validationPath == "" {
		var validationSize = min(1_000_000, len(dataset)/5)
		validation = dataset[:validationSize]
		training = dataset[validationSize:]
	} else {
		validation, err = LoadDataset(config.validationPath, evaluator, parseValidationSample)
		if err != nil {
			return err
		}
		log.Println("Loaded validation", len(validation))
		training = dataset
	}

	var weights = evaluator.StartingWeights()
	log.Println("Num of weights", len(weights))

	var trainer = NewTrainer(training, validation, weights, config.threads)
	err = trainer.Train(config.epochs)
	if err != nil {
		return err
	}

	var wInt = make([]int, len(trainer.weights))
	for i := range wInt {
		wInt[i] = int(math.Round(100 * trainer.weights[i]))
	}
	fmt.Printf("var w = %#v\n", wInt)

	return nil
}
