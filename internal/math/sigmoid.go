package math

import "math"

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func ReverseSigmoid(x float64) float64 {
	return -math.Log(1/x - 1)
}
