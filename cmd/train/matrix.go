package main

type Matrix struct {
	Data []float64
	Rows int
	Cols int
}

func NewMatrix(rows, cols int) Matrix {
	return Matrix{
		Data: make([]float64, rows*cols),
		Rows: rows,
		Cols: cols,
	}
}

func (m *Matrix) Get(row, col int) float64 {
	return m.Data[col*m.Rows+row]
}

func (m *Matrix) Add(row, col int, delta float64) {
	m.Data[col*m.Rows+row] += delta
}

func (m *Matrix) Reset() {
	for i := range m.Data {
		m.Data[i] = 0
	}
}
