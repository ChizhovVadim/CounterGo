package eval

import "fmt"

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
