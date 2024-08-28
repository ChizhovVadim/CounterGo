package ml

import "math"

type IActivationFn interface {
	Sigma(x float64) float64
	SigmaPrime(x float64) float64
}

type IdentityActivation struct{}

func (*IdentityActivation) Sigma(x float64) float64      { return x }
func (*IdentityActivation) SigmaPrime(x float64) float64 { return 1 }

type ReLuActivation struct{}

func (*ReLuActivation) Sigma(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

func (*ReLuActivation) SigmaPrime(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}

type SigmoidActivation struct{}

func (s *SigmoidActivation) Sigma(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func (s *SigmoidActivation) SigmaPrime(x float64) float64 {
	var y = s.Sigma(x)
	return y * (1 - y)
}
