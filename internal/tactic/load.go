package tactic

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func LoadEpd(filePath string) ([]EpdItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []EpdItem
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var test, err = parseEpdTest(line)
		if err != nil {
			log.Println(err)
			continue
		}
		result = append(result, test)
	}
	//err = scanner.Err()
	return result, nil
}

func parseEpdTest(s string) (EpdItem, error) {
	var bmBegin = strings.Index(s, "bm")
	var bmEnd = strings.Index(s, ";")
	var fen = strings.TrimSpace(s[:bmBegin])
	var sBestMoves = strings.Split(s[bmBegin:bmEnd], " ")[1:]

	var p, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return EpdItem{}, err
	}

	var bestMoves []common.Move
	for _, sBestMove := range sBestMoves {
		var move = common.ParseMoveSAN(&p, sBestMove)
		if move == common.MoveEmpty {
			return EpdItem{}, fmt.Errorf("parse move failed %v", s)
		}
		bestMoves = append(bestMoves, move)
	}
	if len(bestMoves) == 0 {
		return EpdItem{}, fmt.Errorf("empty best moves %v", s)
	}

	return EpdItem{
		content:   s,
		position:  p,
		bestMoves: bestMoves,
	}, nil
}
