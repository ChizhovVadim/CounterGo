package main

import (
	"context"
	"log"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

func run(
	ctx context.Context,
	gameConcurrency int,
	tc timeControl,
) error {
	log.Println("arena started")
	defer log.Println("arena finished")

	log.Println("NumCPU", runtime.NumCPU(),
		"GOMAXPROCS", runtime.GOMAXPROCS(0),
		"gameConcurrency", gameConcurrency)

	log.Printf("%+v\n", tc)

	g, ctx := errgroup.WithContext(ctx)

	var gameInfos = make(chan gameInfo)
	var gameResults = make(chan gameResult)

	g.Go(func() error {
		defer close(gameInfos)
		return loadOpenings(ctx, gameInfos)
	})

	g.Go(func() error {
		return showResults(ctx, gameResults)
	})

	var wg = &sync.WaitGroup{}

	for i := 0; i < gameConcurrency; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return playGames(ctx, tc, gameInfos, gameResults)
		})
	}

	g.Go(func() error {
		wg.Wait()
		close(gameResults)
		return nil
	})

	return g.Wait()
}

func playGames(
	ctx context.Context,
	tc timeControl,
	gameInfos <-chan gameInfo,
	gameResults chan<- gameResult,
) error {
	var engineA = newEngineA()
	var engineB = newEngineB()
	for gameInfo := range gameInfos {
		var res, err = playGame(ctx, engineA, engineB, tc, gameInfo)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case gameResults <- res:
		}
	}
	return nil
}
