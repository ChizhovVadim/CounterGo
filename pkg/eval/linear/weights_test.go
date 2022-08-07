package eval

import (
	"fmt"
	"testing"
)

func TestPgn(t *testing.T) {
	for _, info := range infos {
		if info.Size == 0 {
			continue
		}
		if info.Size < 32 {
			fmt.Print(info.Name, " ")
			for i := 0; i < info.Size; i++ {
				fmt.Printf("(%v,%v)",
					w[2*(info.StartIndex+i)]/100,
					w[1+2*(info.StartIndex+i)]/100)
			}
			fmt.Println()
		}
	}
}
