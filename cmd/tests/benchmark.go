package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func runBenchmark(filepath string, evalName string) error {
	logger.Println("benchmark started",
		"filepath", filepath,
		"evalName", evalName)
	defer logger.Println("benchmark finished")

	var tests, err = loadEpd(filepath)
	if err != nil {
		return err
	}
	var eng = newEngine(evalName)
	eng.Prepare()
	benchmark(tests, eng)
	return nil
}

func benchmark(tests []epdItem, eng UciEngine) {
	var ctx = context.Background()
	var start = time.Now()
	var nodes int64
	for i := range tests {
		var test = &tests[i]
		var searchInfo = eng.Search(ctx, common.SearchParams{
			Positions: []common.Position{test.position},
			Limits:    common.LimitsType{Depth: 10},
		})
		nodes += searchInfo.Nodes
	}
	var elapsed = time.Since(start)
	fmt.Println("Time", elapsed)
	fmt.Println("Nodes", nodes)
	fmt.Println("kNPS", nodes/elapsed.Milliseconds())
}
