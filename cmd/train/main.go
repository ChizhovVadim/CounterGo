package main

import (
	"flag"
	"log"
	"runtime"
)

type Config struct {
	trainingPath   string
	validationPath string
	netFolderPath  string
	threads        int
	epochs         int
	searchWeight   float64
	datasetMaxSize int
}

var config Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&config.trainingPath, "td", "", "Path to training dataset")
	flag.StringVar(&config.validationPath, "vd", "", "Path to validation dataset")
	flag.StringVar(&config.netFolderPath, "net", "", "Final NNUE path directory")
	flag.IntVar(&config.threads, "threads", 1, "Number of threads")
	flag.IntVar(&config.epochs, "epochs", 30, "Number of epochs")
	flag.Float64Var(&config.searchWeight, "sw", 0.5, "Weight of search result in training dataset")
	flag.IntVar(&config.datasetMaxSize, "dms", 1000000, "Max size of dataset")
	flag.Parse()

	log.Printf("%+v", config)

	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	var training, validation []Sample

	training, err := LoadDataset2(config.trainingPath)
	if err != nil {
		return err
	}
	log.Println("Loaded dataset", len(training))
	runtime.GC()

	validation, err = LoadDataset(config.validationPath, zurichessParser)
	if err != nil {
		return err
	}
	log.Println("Loaded validation", len(validation))

	var trainer = NewTrainer(training, validation, []int{769, 512, 1}, config.threads)
	//var trainer = NewTrainer(training, validation, []int{769, 256, 16, 1})
	return trainer.Train(config.epochs, config.netFolderPath)
}
