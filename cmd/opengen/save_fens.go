package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func saveFens(ctx context.Context, filepath string, positions <-chan common.Position) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var ticker = time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var totalCount int
	var uniqueCount int
	var repeats = make(map[uint64]struct{})

	var showProgress = func() {
		var uniqueRatio = float64(uniqueCount) / float64(totalCount)
		log.Printf("Total: %v unique: %v, uniqueRatio: %f\n",
			totalCount, uniqueCount, uniqueRatio)
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			showProgress()
		case pos, positionOk := <-positions:
			if !positionOk {
				break LOOP
			}
			totalCount++
			if _, found := repeats[pos.Key]; found {
				continue
			}
			repeats[pos.Key] = struct{}{}

			uniqueCount++

			var fen = pos.String()
			_, err = fmt.Fprintln(file, fen)
			if err != nil {
				return err
			}
		}
	}

	showProgress()
	return nil
}
