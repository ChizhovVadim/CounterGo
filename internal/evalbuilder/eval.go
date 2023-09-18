package evalbuilder

import (
	"fmt"

	counter "github.com/ChizhovVadim/CounterGo/pkg/eval/counter"
	nnue "github.com/ChizhovVadim/CounterGo/pkg/eval/nnue"
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
		case "nnue":
			return nnue.NewDefaultEvaluationService()
		}
		panic(fmt.Errorf("bad eval %v", key))
	}
}
