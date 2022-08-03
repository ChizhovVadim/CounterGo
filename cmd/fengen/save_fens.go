package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type PositionInfo struct {
	position   common.Position
	score      int
	gameResult float32
}

func saveFens(
	ctx context.Context,
	games <-chan []PositionInfo,
	filepath string,
) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var ticker = time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var gameCount int
	var positionCount int

	var showProgress = func() {
		log.Printf("Total %v games, %v positions\n", gameCount, positionCount)
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			showProgress()
		case game, gameOk := <-games:
			if !gameOk {
				break LOOP
			}
			err = writeGame(file, game)
			if err != nil {
				return err
			}
			gameCount++
			positionCount += len(game)
		}
	}

	showProgress()
	return nil
}

func writeGame(w io.Writer, game []PositionInfo) error {
	for i := range game {
		var item = &game[i]
		var fen = item.position.String()
		var score = item.score
		// score from white point of view
		if !item.position.WhiteMove {
			score = -score
		}
		var _, err = fmt.Fprintf(w, "%v;%v;%v\n",
			fen,
			score,
			item.gameResult)
		if err != nil {
			return err
		}
	}
	return nil
}
