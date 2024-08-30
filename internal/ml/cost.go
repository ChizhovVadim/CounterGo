package ml

import (
	"fmt"
	"math"
)

func NewCost(name string) (IModelCost, error) {
	switch name {
	case "mse":
		return &MSECost{}, nil
	case "abs":
		return &AbsCost{}, nil
	}
	return nil, fmt.Errorf("bad cost name %v", name)
}

type IModelCost interface {
	Cost(predicted, target float64) float64
	CostPrime(predicted, target float64) float64
}

type MSECost struct{}

func (*MSECost) Cost(predicted, target float64) float64 {
	var x = predicted - target
	return x * x
}

func (*MSECost) CostPrime(predicted, target float64) float64 {
	return 2 * (predicted - target)
}

type AbsCost struct{}

func (*AbsCost) Cost(predicted, target float64) float64 {
	var x = predicted - target
	if x < 0 {
		return -x
	}
	return x
}

func (*AbsCost) CostPrime(predicted, target float64) float64 {
	var x = predicted - target
	if x < 0 {
		return -1
	}
	return 1
}

type SigmoidMSECost struct{}

func (*SigmoidMSECost) Cost(predicted, target float64) float64 {
	var x = sigmoid(predicted) - target
	return x * x
}

func (*SigmoidMSECost) CostPrime(predicted, target float64) float64 {
	var sigmoid = sigmoid(predicted)
	var sigmoidPrime = sigmoid * (1 - sigmoid)
	return 2 * (sigmoid - target) * sigmoidPrime
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}
