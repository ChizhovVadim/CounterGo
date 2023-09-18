package eval

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func TestWeights(t *testing.T) {
	for _, f := range features {
		if f.Size == 32 && strings.HasSuffix(f.Name, "PST") {
			printPst32(f.Name, w[2*f.Index:2*(f.Index+f.Size)])
		} else {
			fmt.Print(f.Name)
			for i := 0; i < f.Size; i++ {
				fmt.Printf("(%v,%v)",
					w[2*(f.Index+i)]/100, w[2*(f.Index+i)+1]/100)
			}
			fmt.Println()
		}
	}
}

func printPst32(name string, pst []int) {
	fmt.Println(name)
	for i := 0; i < 64; i++ {
		var sq = common.FlipSquare(i)
		var sq32 = relativeSq32(common.SideWhite, sq)
		fmt.Printf("(%v,%v)",
			pst[2*sq32]/100,
			pst[1+2*sq32]/100)
		if common.File(sq) == common.FileH {
			fmt.Println()
		}
	}
}
