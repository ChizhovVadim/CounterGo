package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ChizhovVadim/CounterGo/common"
)

type UciEngine interface {
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

func benchmark(tests []epdItem, eng UciEngine) {
	logger.Println("benchmark started")
	defer logger.Println("benchmark finished")
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
