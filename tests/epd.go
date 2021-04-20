package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ChizhovVadim/CounterGo/common"
)

type epdItem struct {
	content   string
	position  common.Position
	bestMoves []common.Move
}

func loadEpd(filePath string) ([]epdItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []epdItem
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var test, err = parseEpdTest(line)
		if err != nil {
			logger.Println(err)
			continue
		}
		result = append(result, test)
	}
	//err = scanner.Err()
	return result, nil
}

func parseEpdTest(s string) (epdItem, error) {
	var bmBegin = strings.Index(s, "bm")
	var bmEnd = strings.Index(s, ";")
	var fen = strings.TrimSpace(s[:bmBegin])
	var sBestMoves = strings.Split(s[bmBegin:bmEnd], " ")[1:]

	var p, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return epdItem{}, err
	}

	var bestMoves []common.Move
	for _, sBestMove := range sBestMoves {
		var move = common.ParseMoveSAN(&p, sBestMove)
		if move == common.MoveEmpty {
			return epdItem{}, fmt.Errorf("parse move failed %v", s)
		}
		bestMoves = append(bestMoves, move)
	}
	if len(bestMoves) == 0 {
		return epdItem{}, fmt.Errorf("empty best moves %v", s)
	}

	return epdItem{
		content:   s,
		position:  p,
		bestMoves: bestMoves,
	}, nil
}
