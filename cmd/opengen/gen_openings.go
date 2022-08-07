package main

import (
	"context"
	"log"
	"sync"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

func generateOpeningsPipeline(
	ctx context.Context,
	threads int,
	inputPgnFilePath string,
	outputFenFilePath string,
	ply int,
) error {
	log.Println("generateOpenings started")
	defer log.Println("generateOpenings finished")

	g, ctx := errgroup.WithContext(ctx)

	var pgns = make(chan string, 128)
	var positions = make(chan common.Position, 128)

	g.Go(func() error {
		defer close(pgns)
		return pgn.LoadPgns(ctx, inputPgnFilePath, pgns)
	})

	g.Go(func() error {
		return saveFens(ctx, outputFenFilePath, positions)
	})

	var wg = &sync.WaitGroup{}

	for i := 0; i < threads; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return extractOpenings(ctx, ply, pgns, positions)
		})
	}

	g.Go(func() error {
		wg.Wait()
		close(positions)
		return nil
	})

	return g.Wait()
}

func extractOpenings(ctx context.Context, ply int, pgns <-chan string, positions chan<- common.Position) error {
	for spgn := range pgns {
		var game, err = pgn.ParseGame(spgn)
		if err != nil {
			log.Println("Parse game fail", err, spgn)
			continue
		}

		if len(game.Items) <= ply {
			continue
		}
		var pos = game.Items[ply].Position

		select {
		case <-ctx.Done():
			return ctx.Err()
		case positions <- pos:
		}
	}
	return nil
}
