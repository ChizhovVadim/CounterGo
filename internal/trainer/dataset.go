package trainer

import (
	"context"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

type Sample struct {
	Input  []int16
	Target float32
}

func loadDataset(
	ctx context.Context,
	datasetProvider IDatasetProvider,
	mirror bool,
) ([]Sample, error) {
	g, ctx := errgroup.WithContext(ctx)

	var dataset = make(chan domain.DatasetItem, 128)

	g.Go(func() error {
		defer close(dataset)
		return datasetProvider.Load(ctx, dataset)
	})

	var result []Sample

	g.Go(func() error {
		var samples, err = processDataset(ctx, dataset, mirror)
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
	mirror bool,
) ([]Sample, error) {
	var result []Sample
	for item := range dataset {
		pos, err := common.NewPositionFromFEN(item.Fen)
		if err != nil {
			return nil, err
		}
		input := toFeatures(&pos)
		result = append(result, Sample{
			Input:  input,
			Target: float32(item.Target),
		})
		if mirror {
			result = append(result, Sample{
				Input:  mirrorInput(input),
				Target: float32(1 - item.Target),
			})
		}
	}
	return result, nil
}
