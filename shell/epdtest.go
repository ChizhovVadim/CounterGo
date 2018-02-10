package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ChizhovVadim/CounterGo/common"
)

type testItem struct {
	content   string
	position  *common.Position
	bestMoves []common.Move
}

func RunEpdTest(filePath string, uciEngine UciEngine) {
	var epdTests = loadEpdTests(filePath)
	fmt.Printf("Loaded %v tests\n", len(epdTests))
	fmt.Println("Test started...")
	var start = time.Now()
	var total, solved int
	for _, test := range epdTests {
		var searchParams = common.SearchParams{
			Positions: []*common.Position{test.position},
			Limits:    common.LimitsType{MoveTime: 3000},
		}
		var searchResult = uciEngine.Search(searchParams)

		var passed = false
		for _, bm := range test.bestMoves {
			if bm == searchResult.MainLine[0] {
				passed = true
				break
			}
		}

		total++
		if passed {
			solved++
		}

		fmt.Println(test.content)
		PrintSearchInfo(searchResult)

		fmt.Printf("Solved: %v, Total: %v\n", solved, total)
		fmt.Println()
	}
	fmt.Printf("Test finished. Elapsed: %v\n", time.Since(start))
}

func loadEpdTests(filePath string) (result []*testItem) {
	var err = processFileByLines(filePath, func(line string) {
		var test = parseEpdTest(line)
		if test != nil {
			result = append(result, test)
		}
	})
	if err != nil {
		fmt.Println(err)
	}
	return
}

func processFileByLines(filePath string, processor func(line string)) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		processor(line)
	}
	err = scanner.Err()
	return
}

func parseEpdTest(s string) *testItem {
	var bmBegin = strings.Index(s, "bm")
	var bmEnd = strings.Index(s, ";")
	var fen = strings.TrimSpace(s[:bmBegin])
	var p = common.NewPositionFromFEN(fen)
	var sBestMoves = strings.Split(s[bmBegin:bmEnd], " ")[1:]
	var bestMoves []common.Move
	for _, sBestMove := range sBestMoves {
		var move = parseEpdMove(p, sBestMove)
		if move == common.MoveEmpty {
			return nil
		}
		bestMoves = append(bestMoves, move)
	}
	if len(bestMoves) == 0 {
		return nil
	}
	return &testItem{
		content:   s,
		position:  p,
		bestMoves: bestMoves,
	}
}

func parseEpdMove(p *common.Position, s string) common.Move {
	s = strings.TrimRight(s, "+")
	var piece = 2 + strings.Index("NBRQK", s[0:1])
	var to = common.ParseSquare(s[len(s)-2:])
	var ml = common.GenerateLegalMoves(p)
	var moves []common.Move
	for _, move := range ml {
		if move.MovingPiece() == piece &&
			move.To() == to {
			moves = append(moves, move)
		}
	}
	if len(moves) == 1 {
		return moves[0]
	}
	return common.MoveEmpty
}
