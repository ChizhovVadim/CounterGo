package eval

import (
	"sync"
)

func loadWeightsСached(loadFunc func() (*Weights, error)) func() (*Weights, error) {
	var once sync.Once
	var weights *Weights
	var err error
	return func() (*Weights, error) {
		once.Do(func() {
			weights, err = loadFunc()
		})
		return weights, err
	}
}

var loadDefaultWeightsСached = loadWeightsСached(loadDefaultWeights)

// TODO return err
func NewDefaultEvaluationService() *EvaluationService {
	var weights, err = loadDefaultWeightsСached()
	if err != nil {
		panic(err)
	}
	return NewEvaluationService(weights)
}
