package eval

import (
	"fmt"
	"testing"

	. "github.com/ChizhovVadim/CounterGo/common"
)

func TestWeights(t *testing.T) {
	var w = &Weights{}
	w.init()

	fmt.Println("BishopPairMaterial", w.BishopPairMaterial)

	printPst("Pawn", w.PST[SideWhite][Pawn])
	printPst("Knight", w.PST[SideWhite][Knight])
	printPst("Bishop", w.PST[SideWhite][Bishop])
	printPst("Rook", w.PST[SideWhite][Rook])
	printPst("Queen", w.PST[SideWhite][Queen])
	printPst("King", w.PST[SideWhite][King])
}

func printPst(name string, source [64]Score) {
	fmt.Println("PST", name)
	for i := 0; i < 64; i++ {
		var sq = FlipSquare(i)
		fmt.Print(source[sq])
		if File(sq) == FileH {
			fmt.Println()
		}
	}
}
