package eval

import (
	"fmt"
	"testing"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func TestWeights(t *testing.T) {
	for _, info := range infos {
		if info.Size == 0 {
			continue
		}
		if info.Size == 32 {
			printPst32(info.Name, w[2*info.StartIndex:2*(info.StartIndex+info.Size)])
		} else if info.Size < 32 {
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

func TestWeightsGolang(t *testing.T) {
	for _, info := range infos {
		if info.Size == 0 {
			continue
		}
		if info.Size == 1 {
			fmt.Printf("var %v=Score{Mg:%v,Eg:%v}\n",
				info.Name,
				w[2*info.StartIndex],
				w[1+2*info.StartIndex])
		} else {
			fmt.Printf("var %v=[%v]Score{", info.Name, info.Size)
			for i := 0; i < info.Size; i++ {
				fmt.Printf("{Mg:%v,Eg:%v}",
					w[2*(info.StartIndex+i)],
					w[1+2*(info.StartIndex+i)])
				if i < info.Size {
					fmt.Print(",")
				}
			}
			fmt.Printf("}\n")
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
