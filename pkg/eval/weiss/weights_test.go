package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func TestWeights(t *testing.T) {
	var w = &Weights{}
	w.init()

	var enc = json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	enc.Encode(&w)

	printPst("Pawn", w.PSQT[0][common.Pawn])
	printPst("Knight", w.PSQT[0][common.Knight])
	printPst("Bishop", w.PSQT[0][common.Bishop])
	printPst("Rook", w.PSQT[0][common.Rook])
	printPst("Queen", w.PSQT[0][common.Queen])
	printPst("King", w.PSQT[0][common.King])
	//printPst("PawnConnected", w.PawnConnected)
}

func printPst(name string, source [64]Score) {
	fmt.Println("PST", name)
	for i := 0; i < 64; i++ {
		var sq = common.FlipSquare(i)
		fmt.Print(source[sq])
		if common.File(sq) == common.FileH {
			fmt.Println()
		}
	}
}
