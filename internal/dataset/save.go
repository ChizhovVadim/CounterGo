package dataset

import (
	"context"
	"log"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
)

func (dp *DatasetProvider) mergeDataset(
	ctx context.Context,
	input <-chan datasetInfo,
	output chan<- domain.DatasetItem,
	datasetReady chan<- struct{},
) error {
	var repeats = make(map[uint64]struct{})
	var positionCount int
	var repeatCount int

	for info := range input {
		if dp.MaxPosCount != 0 && positionCount >= dp.MaxPosCount {
			if datasetReady != nil {
				close(datasetReady)
				datasetReady = nil
			}
			continue
		}

		if _, found := repeats[info.key]; found {
			repeatCount++
			continue
		}
		repeats[info.key] = struct{}{}
		output <- domain.DatasetItem{
			Fen:    info.fen,
			Target: info.target,
		}
		positionCount++
	}
	log.Println("mergeDataset",
		"positionCount", positionCount,
		"repeatCount", repeatCount)
	return nil
}
