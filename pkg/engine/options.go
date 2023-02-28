package engine

import (
	"math"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Options struct {
	Hash               int
	Threads            int
	ExperimentSettings bool
	ProgressMinNodes   int
	reductions         [64][64]int
}

func NewOptions() Options {
	var result = Options{
		Hash:               16,
		Threads:            1,
		ExperimentSettings: false,
		ProgressMinNodes:   1_000_000,
	}
	result.InitLmr(LmrMult)
	return result
}

func (o *Options) Lmr(d, m int) int {
	return o.reductions[common.Min(d, 63)][common.Min(m, 63)]
}

func (o *Options) InitLmr(f func(d, m float64) float64) {
	initLmr(&o.reductions, f)
}

func initLmr(reductions *[64][64]int,
	f func(d, m float64) float64) {
	for d := 1; d < 64; d++ {
		for m := 1; m < 64; m++ {
			var r = f(float64(d), float64(m))
			reductions[d][m] = int(r)
		}
	}
}

func LmrMult(d, m float64) float64 {
	return lirp(math.Log(d)*math.Log(m), math.Log(5)*math.Log(22), math.Log(63)*math.Log(63), 3, 8)
}

func lirp(x, x1, x2, y1, y2 float64) float64 {
	return y1 + (y2-y1)*(x-x1)/(x2-x1)
}
