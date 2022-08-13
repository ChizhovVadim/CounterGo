package evalbuilder

import (
	"fmt"

	counter "github.com/ChizhovVadim/CounterGo/pkg/eval/counter"
	fast "github.com/ChizhovVadim/CounterGo/pkg/eval/fast"
	linear "github.com/ChizhovVadim/CounterGo/pkg/eval/linear"
	material "github.com/ChizhovVadim/CounterGo/pkg/eval/material"
	pesto "github.com/ChizhovVadim/CounterGo/pkg/eval/pesto"
	weiss "github.com/ChizhovVadim/CounterGo/pkg/eval/weiss"
)

func NewEvalBuilder(key string) func() interface{} {
	return func() interface{} {
		switch key {
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
		}
		panic(fmt.Errorf("bad eval %v", key))
	}
}
