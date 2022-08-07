package main

import "math"

var (
	SigmoidScale = 3.5 / 512
	LearningRate = 0.01
)

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(SigmoidScale*(-x)))
}

func SigmoidPrime(x float64) float64 {
	return x * (1 - x) * SigmoidScale
}

func ValidationCost(output, target float64) float64 {
	var x = output - target
	return x * x
}

func CalculateCostGradient(output, target float64) float64 {
	return 2.0 * (output - target)
}
