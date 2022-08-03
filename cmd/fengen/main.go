package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/internal/quiet"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
	eval "github.com/ChizhovVadim/CounterGo/pkg/eval/counter"
)

type IQuietService interface {
	IsQuiet(p *common.Position) bool
}

func quietServiceBuilder() IQuietService {
	return quiet.NewQuietService(eval.NewEvaluationService(), 30)
}

type IEngine interface {
	Prepare()
	Clear()
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

func engineBuilder() IEngine {
	var eng = engine.NewEngine(func() engine.Evaluator {
		return eval.NewEvaluationService()
	})
	eng.Hash = 32
	eng.Threads = 1
	eng.Prepare()
	return eng
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
	curUser, err := user.Current()
	if err != nil {
		return err
	}
	homeDir := curUser.HomeDir
	if homeDir == "" {
		return fmt.Errorf("current user home dir empty")
	}

	var chessDir = filepath.Join(homeDir, "chess")

	var settings = Settings{
		GamesFolder: filepath.Join(chessDir, "pgn"),
		ResultPath:  filepath.Join(chessDir, "fengen.txt"),
		Threads:     max(1, runtime.NumCPU()/2),
	}

	flag.StringVar(&settings.GamesFolder, "input", settings.GamesFolder, "Path to folder with PGN files")
	flag.StringVar(&settings.ResultPath, "output", settings.ResultPath, "Path to output fen file")
	flag.IntVar(&settings.Threads, "threads", settings.Threads, "Number of threads")
	flag.Parse()

	log.Printf("%+v", settings)

	pgnFiles, err := pgnFiles(settings.GamesFolder)
	if err != nil {
		return err
	}
	if len(pgnFiles) == 0 {
		return fmt.Errorf("At least one PGN file is expected")
	}

	var datasetService = &DatasetService1{
		quietServiceBuilder: quietServiceBuilder,
		threads:             settings.Threads,
		pgnFiles:            pgnFiles,
		resultPath:          settings.ResultPath,
	}
	return datasetService.Run(context.Background())
}
