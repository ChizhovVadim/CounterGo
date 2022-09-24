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

	var repeats = make(map[uint64]struct{})

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
			n, err := writeGame(file, game, repeats)
			if err != nil {
				return err
			}
			gameCount++
			positionCount += n
		}
	}

	showProgress()
	return nil
}

func writeGame(w io.Writer, game []PositionInfo, repeats map[uint64]struct{}) (int, error) {
	var n = 0
	for i := range game {
		var item = &game[i]
		if _, found := repeats[item.position.Key]; found {
			continue
		}
		repeats[item.position.Key] = struct{}{}

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
			return n, err
		}
		n++
	}
	return n, nil
}
