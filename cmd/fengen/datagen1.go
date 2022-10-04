package main

import (
	"context"
	"log"
	"sync"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"

	"golang.org/x/sync/errgroup"
)

type DatasetService1 struct {
	quietServiceBuilder func() IQuietService
	threads             int
	pgnFiles            []string
	resultPath          string
}

func (ds *DatasetService1) Run(ctx context.Context) error {
	log.Println("fengen started")
	defer log.Println("fengen finished")

	g, ctx := errgroup.WithContext(ctx)

	var pgns = make(chan string, 128)
	var games = make(chan []PositionInfo, 128)

	g.Go(func() error {
		defer close(pgns)
		return pgn.LoadPgnsManyFiles(ctx, ds.pgnFiles, pgns)
	})

	g.Go(func() error {
		return saveFens(ctx, games, ds.resultPath)
	})

	var wg = &sync.WaitGroup{}

	for i := 0; i < ds.threads; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return ds.analyzeGames(ctx, ds.quietServiceBuilder(), pgns, games)
		})
	}

	g.Go(func() error {
		wg.Wait()
		close(games)
		return nil
	})

	return g.Wait()
}

func (ds *DatasetService1) analyzeGames(
	ctx context.Context,
	quietService IQuietService,
	pgns <-chan string,
	games chan<- []PositionInfo,
) error {
	for pgn := range pgns {
		var game, err = ds.analyzeGame(quietService, pgn)
		if err != nil {
			log.Println("AnalyzeGame error", err, pgn)
			continue
		}
		if len(game) != 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case games <- game:
			}
		}
	}
	return nil
}

func (ds *DatasetService1) analyzeGame(
	quietService IQuietService,
	sPgn string,
) ([]PositionInfo, error) {
	var game, err = pgn.ParseGame(sPgn)
	if err != nil {
		return nil, err
	}

	gameResult, err := calcGameResult(&game)
	if err != nil {
		if err == errGameResultNone {
			return nil, nil
		}
		return nil, err
	}

	var repeatPositions = make(map[uint64]struct{})
	var result []PositionInfo

	for i := range game.Items {
		if i > 0 {
			repeatPositions[game.Items[i-1].Position.Key] = struct{}{}
		}

		var item = &game.Items[i]

		var curPosition = &item.Position
		var score int

		var comment = item.Comment
		if comment.Depth < 8 ||
			comment.Score.Mate != 0 {
			continue
		}
		score = comment.Score.Centipawns

		if curPosition.IsCheck() {
			continue
		}
		if _, found := repeatPositions[curPosition.Key]; found {
			continue
		}
		if !quietService.IsQuiet(curPosition) {
			continue
		}

		result = append(result, PositionInfo{
			position:   item.Position,
			score:      score,
			gameResult: gameResult,
		})
	}

	return result, nil
}
