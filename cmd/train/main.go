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
	datasetMaxSize int
}

var config Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&config.trainingPath, "td", "", "Path to training dataset")
	flag.StringVar(&config.validationPath, "vd", "", "Path to validation dataset")
	flag.StringVar(&config.netFolderPath, "net", "", "Final NNUE path directory")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU(), "Number of threads")
	flag.IntVar(&config.epochs, "epochs", 30, "Number of epochs")
	flag.IntVar(&config.datasetMaxSize, "dms", 5_000_000, "Max size of dataset")
	flag.Parse()

	log.Printf("%+v", config)

	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	dataset, err := LoadDataset(config.trainingPath)
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
		validation, err = LoadZurichessDataset(config.validationPath)
		if err != nil {
			return err
		}
		log.Println("Loaded validation", len(validation))
		training = dataset
	}

	var trainer = NewTrainer(training, validation, []int{769, 512, 1}, config.threads, 0)
	return trainer.Train(config.epochs, config.netFolderPath)
}
