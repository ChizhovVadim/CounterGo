package main

import "math"

type ActivationFn interface {
	Sigma(x float64) float64
	SigmaPrime(x float64) float64
}

type ReLu struct{}

func (s *ReLu) Sigma(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

func (s *ReLu) SigmaPrime(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}

type Sigmoid struct {
	SigmoidScale float64
}

func (s *Sigmoid) Sigma(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(s.SigmoidScale*(-x)))
}

func (s *Sigmoid) SigmaPrime(x float64) float64 {
	var y = s.Sigma(x)
	return y * (1 - y) * s.SigmoidScale
}
