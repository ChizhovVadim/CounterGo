package evalbuilder

import (
	"fmt"

	counter "github.com/ChizhovVadim/CounterGo/pkg/eval/counter"
	fast "github.com/ChizhovVadim/CounterGo/pkg/eval/fast"
	linear "github.com/ChizhovVadim/CounterGo/pkg/eval/linear"
	material "github.com/ChizhovVadim/CounterGo/pkg/eval/material"
	nnue "github.com/ChizhovVadim/CounterGo/pkg/eval/nnue"
	pesto "github.com/ChizhovVadim/CounterGo/pkg/eval/pesto"
	weiss "github.com/ChizhovVadim/CounterGo/pkg/eval/weiss"
)

func Get(key string) func() interface{} {
	return func() interface{} {
		switch key {
		case "":
			if nnue.AvxInstructions {
				return nnue.NewDefaultEvaluationService()
			} else {
				return counter.NewEvaluationService()
			}
		case "counter":
			return counter.NewEvaluationService()
		case "weiss":
			return weiss.NewEvaluationService()
		case "linear":
			return linear.NewEvaluationService()
		case "pesto":
			return pesto.NewEvaluationService()
		case "material":
			return material.NewEvaluationService()
		case "fast":
			return fast.NewEvaluationService()
		case "nnue":
			return nnue.NewDefaultEvaluationService()
		}
		panic(fmt.Errorf("bad eval %v", key))
	}
}
