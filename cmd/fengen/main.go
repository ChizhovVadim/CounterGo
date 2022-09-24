package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ChizhovVadim/CounterGo/internal/quiet"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	eval "github.com/ChizhovVadim/CounterGo/pkg/eval/counter"
)

type IQuietService interface {
	IsQuiet(p *common.Position) bool
}

func quietServiceBuilder() IQuietService {
	return quiet.NewQuietService(eval.NewEvaluationService(), 30)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

type Settings struct {
	GamesFolder string
	ResultPath  string
	Threads     int
}

func run() error {
	var settings = Settings{
		Threads: 1,
	}

	flag.StringVar(&settings.GamesFolder, "input", "", "Path to folder with PGN files")
	flag.StringVar(&settings.ResultPath, "output", "", "Path to output fen file")
	flag.IntVar(&settings.Threads, "threads", settings.Threads, "Number of threads")
	flag.Parse()

	log.Printf("%+v", settings)

	pgnFiles, err := pgnFiles(settings.GamesFolder)
	if err != nil {
		return err
	}
	if len(pgnFiles) == 0 {
		return fmt.Errorf("at least one PGN file is expected")
	}

	var datasetService = &DatasetService1{
		quietServiceBuilder: quietServiceBuilder,
		threads:             settings.Threads,
		pgnFiles:            pgnFiles,
		resultPath:          settings.ResultPath,
	}
	return datasetService.Run(context.Background())
}
