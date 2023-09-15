package tuner

import (
	"context"

	"github.com/ChizhovVadim/CounterGo/internal/domain"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

type Sample struct {
	Target float32
	domain.TuneEntry
}

func loadDataset(
	ctx context.Context,
	datasetProvider IDatasetProvider,
	e ITunableEvaluator,
) ([]Sample, error) {

	g, ctx := errgroup.WithContext(ctx)

	var dataset = make(chan domain.DatasetItem, 128)

	g.Go(func() error {
		defer close(dataset)
		return datasetProvider.Load(ctx, dataset)
	})

	var result []Sample

	g.Go(func() error {
		var samples, err = processDataset(ctx, dataset, e)
		if err != nil {
			return err
		}
		result = samples
		return nil
	})

	var err = g.Wait()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func processDataset(
	ctx context.Context,
	dataset <-chan domain.DatasetItem,
	e ITunableEvaluator,
) ([]Sample, error) {
	var result []Sample
	e.EnableTuning()
	for item := range dataset {
		var pos, err = common.NewPositionFromFEN(item.Fen)
		if err != nil {
			return nil, err
		}
		result = append(result, Sample{
			Target:    float32(item.Target),
			TuneEntry: e.ComputeFeatures(&pos),
		})
	}
	return result, nil
}
