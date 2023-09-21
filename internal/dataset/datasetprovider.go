package dataset

import (
	"context"
	"log"
	"sync"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/internal/pgn"

	"golang.org/x/sync/errgroup"
)

type datasetInfo struct {
	fen    string
	key    uint64
	target float64
}

type DatasetProvider struct {
	SigmoidScale float64
	MaxPosCount  int
	GamesFolder  string
	Threads      int
	SearchRatio  float64
}

func (dp *DatasetProvider) Load(
	ctx context.Context,
	dataset chan<- domain.DatasetItem,
) error {
	log.Println("load dataset started")
	defer log.Println("load dataset finished")

	g, ctx := errgroup.WithContext(ctx)

	var games = make(chan pgn.GameRaw, 128)
	var results = make(chan datasetInfo, 128)
	var datasetReady = make(chan struct{})

	g.Go(func() error {
		defer close(games)
		return dp.loadGames(ctx, games, datasetReady)
	})

	g.Go(func() error {
		return dp.mergeDataset(ctx, results, dataset, datasetReady)
	})

	var wg = &sync.WaitGroup{}
	for i := 0; i < dp.Threads; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return dp.analyzeGames(ctx, games, results)
		})
	}

	g.Go(func() error {
		wg.Wait()
		close(results)
		return nil
	})

	return g.Wait()
}
