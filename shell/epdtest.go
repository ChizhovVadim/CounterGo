package shell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ChizhovVadim/CounterGo/common"
)

type testItem struct {
	content   string
	position  common.Position
	bestMoves []common.Move
}

func RunEpdTest(epdTests []testItem, uciEngine UciEngine, moveTime int) {
	fmt.Println("Test started...")
	var start = time.Now()
	var total, solved int
	for _, test := range epdTests {
		var searchParams = common.SearchParams{
			Positions: []common.Position{test.position},
			Limits:    common.LimitsType{MoveTime: moveTime},
		}
		var searchResult = uciEngine.Search(context.Background(), searchParams)

		var passed = false
		for _, bm := range test.bestMoves {
			if bm == searchResult.MainLine[0] {
				passed = true
				break
			}
		}

		total++
		if passed {
			solved++
		}

		fmt.Println(test.content)
		PrintSearchInfo(searchResult)

		fmt.Printf("Solved: %v, Total: %v\n", solved, total)
		fmt.Println()
	}
	fmt.Printf("Test finished. Elapsed: %v\n", time.Since(start))
}

func LoadEpdTests(filePath string) ([]testItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []testItem
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var test, err = parseEpdTest(line)
		if err != nil {
			fmt.Println(err)
		} else {
			result = append(result, test)
		}
	}
	//err = scanner.Err()
	return result, nil
}

func parseEpdTest(s string) (testItem, error) {
	var bmBegin = strings.Index(s, "bm")
	var bmEnd = strings.Index(s, ";")
	var fen = strings.TrimSpace(s[:bmBegin])
	var p, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return testItem{}, err
	}
	var sBestMoves = strings.Split(s[bmBegin:bmEnd], " ")[1:]
	var bestMoves []common.Move
	for _, sBestMove := range sBestMoves {
		var move = common.ParseMoveSAN(&p, sBestMove)
		if move == common.MoveEmpty {
			return testItem{}, fmt.Errorf("parse move failed %v", s)
		}
		bestMoves = append(bestMoves, move)
	}
	if len(bestMoves) == 0 {
		return testItem{}, fmt.Errorf("empty best moves %v", s)
	}
	return testItem{
		content:   s,
		position:  p,
		bestMoves: bestMoves,
	}, nil
}
