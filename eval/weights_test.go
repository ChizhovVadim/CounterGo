package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ChizhovVadim/CounterGo/common"
)

func TestWeights(t *testing.T) {
	var w = &Weights{}
	w.init()

	var enc = json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	enc.Encode(&w)

	printPst("Knight", w.PST[0][common.Knight])
	printPst("Queen", w.PST[0][common.Queen])
	printPst("King", w.PST[0][common.King])
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
