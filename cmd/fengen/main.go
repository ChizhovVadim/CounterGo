package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"
	"runtime"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

type Settings struct {
	GamesFolder                 string
	ResultPath                  string
	Threads                     int
	SearchRatio                 float64
	CheckNoisyOnlyForSideToMove bool
}

func run() error {
	var settings = Settings{
		CheckNoisyOnlyForSideToMove: true,
	}

	flag.StringVar(&settings.GamesFolder, "input", "", "Path to folder with PGN files")
	flag.StringVar(&settings.ResultPath, "output", "", "Path to output fen file")
	flag.IntVar(&settings.Threads, "threads", runtime.NumCPU(), "Number of threads")
	flag.Float64Var(&settings.SearchRatio, "sw", 1, "Weight of search result in training dataset")
	flag.Parse()

	if settings.ResultPath == "" {
		settings.ResultPath = filepath.Join(settings.GamesFolder, "dataset.txt")
	}

	log.Printf("%+v", settings)

	return generateDataset(context.Background(), settings)
}
