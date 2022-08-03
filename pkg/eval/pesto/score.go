package eval

import (
	"fmt"
)

type Score int32

func (s Score) Middle() int16 {
	return int16(uint32(s+0x8000) >> 16)
}

func (s Score) End() int16 {
	return int16(s)
}

func S(middle, end int16) Score {
	return Score((uint32(middle) << 16)) + Score(end)
}

func (s Score) String() string {
	return fmt.Sprintf("Score(%d, %d)", s.Middle(), s.End())
}
