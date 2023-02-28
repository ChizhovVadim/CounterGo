package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func runSolveTactic(filepath string, evalName string, moveTime time.Duration) error {
	logger.Println("solveTactic started",
		"filepath", filepath,
		"evalName", evalName,
		"moveTime", moveTime)
	defer logger.Println("solveTactic finished")

	var tests, err = loadEpd(filepath)
	if err != nil {
		return err
	}
	var eng = newEngine(evalName)
	eng.Options.ProgressMinNodes = 0
	eng.Prepare()
	solveTactic(tests, eng, moveTime)
	return nil
}

func solveTactic(tests []epdItem, eng UciEngine, moveTime time.Duration) error {
	var start = time.Now()
	var total, solved int
	for _, test := range tests {
		var searchResult = executeTest(test, eng, moveTime)
		var passed = isTestPassed(test, searchResult.MainLine[0])

		total++
		if passed {
			solved++
		}

		fmt.Println(test.content)
		fmt.Printf("%+v\n", searchResult)
		fmt.Printf("Solved: %v, Total: %v\n", solved, total)
		fmt.Println()
	}
	logger.Printf("Test finished. Elapsed: %v\n", time.Since(start))
	return nil
}

func cancelSearch(test epdItem, cancel context.CancelFunc) func(si common.SearchInfo) {
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

func isTestPassed(test epdItem, foundMove common.Move) bool {
	for _, bm := range test.bestMoves {
		if bm == foundMove {
			return true
		}
	}
	return false
}

func executeTest(test epdItem, uciEngine UciEngine, moveTime time.Duration) common.SearchInfo {
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var searchParams = common.SearchParams{
		Positions: []common.Position{test.position},
		Limits:    common.LimitsType{MoveTime: int(moveTime.Milliseconds())},
		Progress:  cancelSearch(test, cancel),
	}
	return uciEngine.Search(ctx, searchParams)
}
