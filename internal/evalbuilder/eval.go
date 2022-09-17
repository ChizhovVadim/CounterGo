package evalbuilder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	counter "github.com/ChizhovVadim/CounterGo/pkg/eval/counter"
	fast "github.com/ChizhovVadim/CounterGo/pkg/eval/fast"
	linear "github.com/ChizhovVadim/CounterGo/pkg/eval/linear"
	material "github.com/ChizhovVadim/CounterGo/pkg/eval/material"
	nnue "github.com/ChizhovVadim/CounterGo/pkg/eval/nnue"
	pesto "github.com/ChizhovVadim/CounterGo/pkg/eval/pesto"
	weiss "github.com/ChizhovVadim/CounterGo/pkg/eval/weiss"
)

var once sync.Once
var weights *nnue.Weights

func Build(key string) interface{} {
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
	case "nnue":
		once.Do(func() {
			var w, err = loadWeights()
			if err != nil {
				panic(err)
			}
			weights = w
		})
		return nnue.NewEvaluationService(weights)
	}
	panic(fmt.Errorf("bad eval %v", key))
}

func loadWeights() (*nnue.Weights, error) {
	var exePath, err = os.Executable()
	if err != nil {
		return nil, err
	}
	var filePath = filepath.Join(filepath.Dir(exePath), "default.nn")
	weights, err = nnue.LoadWeights(filePath)
	if err != nil {
		return nil, err
	}
	log.Printf("Loaded nnue file %v", filePath)
	return weights, nil
}
