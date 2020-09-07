package eval

import (
	"fmt"
	"math"
)

type Score struct {
	Mg int
	Eg int
}

func (s Score) String() string {
	return fmt.Sprintf("Score(%d, %d)", s.Mg, s.Eg)
}

func (s *Score) add(v Score) {
	s.Mg += v.Mg
	s.Eg += v.Eg
}

func (s *Score) sub(v Score) {
	s.Mg -= v.Mg
	s.Eg -= v.Eg
}

func (s *Score) addN(v Score, n int) {
	s.Mg += v.Mg * n
	s.Eg += v.Eg * n
}

func negScore(s Score) Score {
	return Score{-s.Mg, -s.Eg}
}

func makeScore(mg, eg float64) Score {
	return Score{
		Mg: int(math.Round(mg)),
		Eg: int(math.Round(eg)),
	}
}
