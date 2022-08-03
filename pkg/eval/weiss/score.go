package eval

import (
	"encoding/json"
	"fmt"
)

type Score int32

func (s Score) Middle() int {
	return int(int16(uint32(s+0x8000) >> 16))
}

func (s Score) End() int {
	return int(int16(s))
}

func S(middle, end int) Score {
	return Score((uint32(int16(middle)) << 16)) + Score(int16(end))
}

func (s Score) String() string {
	return fmt.Sprintf("Score(%d, %d)", s.Middle(), s.End())
}

func (s *Score) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Mg int `json:"mg"`
		Eg int `json:"eg"`
	}{
		Mg: int(s.Middle()),
		Eg: int(s.End()),
	})
}
