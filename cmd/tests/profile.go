package main

import (
	"context"
	"os"
	"runtime/pprof"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

//go tool pprof cpu.prof
func runProfile(cpuprofile, evalName string) error {
	logger.Println("runProfile started",
		"cpuprofile", cpuprofile,
		"evalName", evalName)
	defer logger.Println("runProfile finished")

	f, err := os.Create(cpuprofile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		return err
	}
	defer pprof.StopCPUProfile()

	var eng = newEngine(evalName)
	pos, err := common.NewPositionFromFEN("r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4")
	if err != nil {
		return err
	}
	eng.Search(context.Background(), common.SearchParams{
		Positions: []common.Position{pos},
		Limits:    common.LimitsType{MoveTime: 5_000},
		Progress:  func(si common.SearchInfo) {},
	})
	return nil
}
