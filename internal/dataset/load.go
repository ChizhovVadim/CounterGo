package dataset

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"
)

func (dp *DatasetProvider) loadGames(
	ctx context.Context,
	games chan<- pgn.GameRaw,
	datasetReady <-chan struct{},
) error {
	pgnFiles, err := pgnFiles(dp.GamesFolder)
	if err != nil {
		return err
	}
	if len(pgnFiles) == 0 {
		return fmt.Errorf("at least one PGN file is expected")
	}
	var gamesCount int
	var errDatasetReady = errors.New("dataset ready")
	for _, filepath := range pgnFiles {
		log.Println("loadGames",
			"filepath", filepath)
		var err = pgn.WalkPgnFile(filepath, func(gr pgn.GameRaw) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-datasetReady:
				return errDatasetReady
			case games <- gr:
				gamesCount++
				return nil
			}
		})
		if err != nil {
			if errors.Is(err, errDatasetReady) {
				break
			}
			return err
		}
	}
	log.Println("loadGames",
		"gamesCount", gamesCount)
	return nil
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
