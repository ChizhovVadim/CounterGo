package tactic

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type IEngine interface {
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

func SolveTactic(tests []EpdItem, eng IEngine, moveTime time.Duration) error {
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
	log.Printf("Test finished. Elapsed: %v\n", time.Since(start))
	return nil
}

func cancelSearch(test EpdItem, cancel context.CancelFunc) func(si common.SearchInfo) {
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

func isTestPassed(test EpdItem, foundMove common.Move) bool {
	for _, bm := range test.bestMoves {
		if bm == foundMove {
			return true
		}
	}
	return false
}

func executeTest(test EpdItem, uciEngine IEngine, moveTime time.Duration) common.SearchInfo {
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var searchParams = common.SearchParams{
		Positions: []common.Position{test.position},
		Limits:    common.LimitsType{MoveTime: int(moveTime.Milliseconds())},
		Progress:  cancelSearch(test, cancel),
	}
	return uciEngine.Search(ctx, searchParams)
}
