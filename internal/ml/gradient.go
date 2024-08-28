package ml

import "math"

const (
	LearningRate = 0.01
	Beta1        = 0.9
	Beta2        = 0.999
)

type Gradient struct {
	Value float64
	M1    float64
	M2    float64
}

type Gradients struct {
	Data []Gradient
	Rows int
	Cols int
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

func NewGradients(rows, cols int) Gradients {
	return Gradients{
		Data: make([]Gradient, cols*rows),
		Rows: rows,
		Cols: cols,
	}
}

func (g *Gradients) Add(row, col int, delta float64) {
	g.Data[col*g.Rows+row].Value += delta
}

func (g *Gradients) AddMatrix(m *Matrix) {
	for i := range g.Data {
		g.Data[i].Value += m.Data[i]
	}
}

func (g *Gradients) AddTo(parent *Gradients) {
	for i := range g.Data {
		parent.Data[i].Value += g.Data[i].Value
		g.Data[i].Value = 0
	}
}

func (g *Gradients) Apply(m *Matrix) {
	for i := range g.Data {
		m.Data[i] -= g.Data[i].Calculate()
		g.Data[i].Value = 0
	}
}
