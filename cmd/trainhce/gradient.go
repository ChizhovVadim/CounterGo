package main

import (
	"math"
)

type Gradient struct {
	Value float64
	M1    float64
	M2    float64
}

const (
	Beta1 float64 = 0.9
	Beta2 float64 = 0.999
)

// Implementing Gradient

func (g *Gradient) Update(delta float64) {
	g.Value += delta
}

func (g *Gradient) Calculate() float64 {

	if g.Value == 0 {
		// nothing to calculate
		return 0
	}

	g.M1 = g.M1*Beta1 + g.Value*(1-Beta1)
	g.M2 = g.M2*Beta2 + (g.Value*g.Value)*(1-Beta2)

	return LearningRate * g.M1 / (math.Sqrt(g.M2) + 1e-8)
}

func (g *Gradient) Reset() {
	g.Value = 0.0
}

func (g *Gradient) Apply(elem *float64) {
	*elem -= g.Calculate()
	g.Reset()
}
