package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

func generateDataset(ctx context.Context, settings Settings) error {
	log.Println("generate started")
	defer log.Println("generate finished")

	g, ctx := errgroup.WithContext(ctx)

	var games = make(chan pgn.GameRaw, 128)
	var results = make(chan DatasetInfo, 128)

	g.Go(func() error {
		defer close(games)
		return loadGames(ctx, settings.GamesFolder, games, 1_000_000)
	})

	g.Go(func() error {
		return saveDataset(ctx, results, settings.ResultPath)
	})

	var wg = &sync.WaitGroup{}
	for i := 0; i < settings.Threads; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return analyzeGames(ctx, games, results, &settings)
		})
	}

	g.Go(func() error {
		wg.Wait()
		close(results)
		return nil
	})

	return g.Wait()
}

func pgnFiles(folderPath string) ([]string, error) {
	dirs, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, de := range dirs {
		if !de.IsDir() && filepath.Ext(de.Name()) == ".pgn" {
			result = append(result, filepath.Join(folderPath, de.Name()))
		}
	}
	return result, nil
}

type DatasetInfo struct {
	fen    string
	key    uint64
	target float64
}

func loadGames(
	ctx context.Context,
	inputFolder string,
	games chan<- pgn.GameRaw,
	maxSize int,
) error {
	pgnFiles, err := pgnFiles(inputFolder)
	if err != nil {
		return err
	}
	if len(pgnFiles) == 0 {
		return fmt.Errorf("at least one PGN file is expected")
	}
	var gamesCount int
	var errMaxSize = errors.New("max size")
	for _, filepath := range pgnFiles {
		log.Println("loadGames",
			"filepath", filepath)
		var err = pgn.WalkPgnFile(filepath, func(gr pgn.GameRaw) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case games <- gr:
				gamesCount++
				if maxSize != 0 && gamesCount >= maxSize {
					return errMaxSize
				}
				return nil
			}
		})
		if err != nil {
			if err == errMaxSize {
				break
			}
			return err
		}
	}
	log.Println("loadGames",
		"gamesCount", gamesCount)
	return nil
}

func saveDataset(
	ctx context.Context,
	dataset <-chan DatasetInfo,
	filepath string,
) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var repeats = make(map[uint64]struct{})
	var positionCount int
	var repeatCount int

	for info := range dataset {
		if _, found := repeats[info.key]; found {
			repeatCount++
			continue
		}
		repeats[info.key] = struct{}{}
		var _, err = fmt.Fprintf(file, "%v;%v\n",
			info.fen,
			info.target)
		if err != nil {
			return err
		}
		positionCount++
	}
	log.Println("saveDataset",
		"filepath", filepath,
		"positionCount", positionCount,
		"repeatCount", repeatCount)
	return nil
}

func analyzeGames(
	ctx context.Context,
	games <-chan pgn.GameRaw,
	dataset chan<- DatasetInfo,
	settings *Settings,
) error {
	for game := range games {
		var err = analyzeGame(ctx, game, dataset, settings)
		if err != nil {
			log.Println("analyzeGame failed",
				"Tags", game.Tags,
				"BodyRaw", game.BodyRaw,
				"err", err)
		}
	}
	return nil
}

func analyzeGame(
	ctx context.Context,
	gameRaw pgn.GameRaw,
	dataset chan<- DatasetInfo,
	settings *Settings,
) error {
	var game, err = pgn.ParseGame(gameRaw)
	if err != nil {
		return err
	}

	var gameResult, gameResOk = calcGameResult(game.Result)
	if !gameResOk {
		return fmt.Errorf("bad game result %v", game.Result)
	}

	var startFen = game.Fen
	if startFen == "" {
		startFen = common.InitialPositionFen
	}
	pos, err := common.NewPositionFromFEN(startFen)
	if err != nil {
		return err
	}

	var repeatPositions = make(map[uint64]struct{})

	for i := range game.Items {
		repeatPositions[pos.Key] = struct{}{}

		//Make move
		var child common.Position
		if !pos.MakeMove(game.Items[i].Move, &child) {
			break
		}
		pos = child

		//filter quiet positions
		var comment = game.Items[i].Comment
		if comment.Depth < 8 {
			continue
		}
		if comment.Score.Mate != 0 {
			continue
		}
		if pos.IsCheck() {
			continue
		}
		if isDraw(&pos) {
			continue
		}
		if _, found := repeatPositions[pos.Key]; found {
			continue
		}
		if isNoisyPos(&pos, settings.CheckNoisyOnlyForSideToMove) {
			continue
		}

		var targetBySearch = sigmoid(float64(comment.Score.Centipawns))
		if pos.WhiteMove {
			//!
			targetBySearch = 1 - targetBySearch
		}

		var target = targetBySearch*settings.SearchRatio + gameResult*(1-settings.SearchRatio)

		//save position
		dataset <- DatasetInfo{
			fen:    pos.String(),
			key:    pos.Key,
			target: target,
		}
	}
	return nil
}

func sigmoid(x float64) float64 {
	const SigmoidScale = 3.5 / 512
	return 1.0 / (1.0 + math.Exp(SigmoidScale*(-x)))
}

func calcGameResult(sGameResult string) (float64, bool) {
	switch sGameResult {
	case pgn.GameResultWhiteWin:
		return 1, true
	case pgn.GameResultBlackWin:
		return 0, true
	case pgn.GameResultDraw:
		return 0.5, true
	default:
		return 0, false
	}
}
