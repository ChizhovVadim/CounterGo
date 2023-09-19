package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func benchmarkHandler() error {
	var evalName = cliArgs.GetString("eval", "")

	log.Println("benchmark started",
		"evalName", evalName)
	defer log.Println("benchmark finished")

	var tests, err = loadEpd(tacticTestsPath)
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
