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

type TestItem struct {
	content   string
	position  common.Position
	bestMoves []common.Move
}

func RunEpdTest(epdTests []TestItem, uciEngine UciEngine, moveTime int) {
	fmt.Println("Test started...")
	var start = time.Now()
	var total, solved int
	for _, test := range epdTests {
		var searchResult = executeTest(test, uciEngine, moveTime)
		var passed = isTestPassed(test, searchResult.MainLine[0])

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

func cancelSearch(test TestItem, cancel context.CancelFunc) func(si common.SearchInfo) {
	var count = 0
	return func(si common.SearchInfo) {
		if isTestPassed(test, si.MainLine[0]) {
			count++
		} else {
			count = 0
		}
		if count >= 3 {
			cancel()
		}
	}
}

func isTestPassed(test TestItem, foundMove common.Move) bool {
	for _, bm := range test.bestMoves {
		if bm == foundMove {
			return true
		}
	}
	return false
}

func executeTest(test TestItem, uciEngine UciEngine, moveTime int) common.SearchInfo {
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var searchParams = common.SearchParams{
		Positions: []common.Position{test.position},
		Limits:    common.LimitsType{MoveTime: moveTime},
		Progress:  cancelSearch(test, cancel),
	}
	return uciEngine.Search(ctx, searchParams)
}

func LoadEpdTests(filePath string) ([]TestItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []TestItem
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

func parseEpdTest(s string) (TestItem, error) {
	var bmBegin = strings.Index(s, "bm")
	var bmEnd = strings.Index(s, ";")
	var fen = strings.TrimSpace(s[:bmBegin])
	var p, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return TestItem{}, err
	}
	var sBestMoves = strings.Split(s[bmBegin:bmEnd], " ")[1:]
	var bestMoves []common.Move
	for _, sBestMove := range sBestMoves {
		var move = common.ParseMoveSAN(&p, sBestMove)
		if move == common.MoveEmpty {
			return TestItem{}, fmt.Errorf("parse move failed %v", s)
		}
		bestMoves = append(bestMoves, move)
	}
	if len(bestMoves) == 0 {
		return TestItem{}, fmt.Errorf("empty best moves %v", s)
	}
	return TestItem{
		content:   s,
		position:  p,
		bestMoves: bestMoves,
	}, nil
}
