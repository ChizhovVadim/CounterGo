package eval

import "fmt"

/*type Score struct {
	mg int
	eg int
}

func S(mg, eg int) Score {
	return Score{mg: mg, eg: eg}
}

func (s Score) Mg() int {
	return s.mg
}

func (s Score) Eg() int {
	return s.eg
}*/

type Score int64

func (s Score) Mg() int {
	return int(int32((s + 1<<31) >> 32))
}

func (s Score) Eg() int {
	return int(int32(s))
}

func S(middle, end int) Score {
	return Score(middle)<<32 + Score(end)
}

func (s Score) String() string {
	return fmt.Sprintf("Score(%d, %d)", s.Mg(), s.Eg())
}
