package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ChizhovVadim/CounterGo/common"
)

type TestItem struct {
	Content   string
	Position  *common.Position
	BestMoves []common.Move
}

func RunEpdTest(filePath string, uciEngine UciEngine) {
	var epdTests = LoadEpdTests(filePath)
	fmt.Printf("Loaded %v tests\n", len(epdTests))
	fmt.Println("Test started...")
	var start = time.Now()
	var total, solved int
	for _, test := range epdTests {
		var searchParams = common.SearchParams{
			Positions: []*common.Position{test.Position},
			Limits:    common.LimitsType{MoveTime: 3000},
		}
		var searchResult = uciEngine.Search(searchParams)

		var passed = false
		for _, bm := range test.BestMoves {
			if bm == searchResult.MainLine[0] {
				passed = true
				break
			}
		}

		total++
		if passed {
			solved++
		}

		fmt.Println(test.Content)
		PrintSearchInfo(searchResult)

		fmt.Printf("Solved: %v, Total: %v\n", solved, total)
		fmt.Println()
	}
	fmt.Printf("Test finished. Elapsed: %v\n", time.Since(start))
}

func LoadEpdTests(filePath string) (result []*TestItem) {
	var err = ProcessFileByLines(filePath, func(line string) {
		var test = ParseEpdTest(line)
		if test != nil {
			result = append(result, test)
		}
	})
	if err != nil {
		fmt.Println(err)
	}
	return
}

func ProcessFileByLines(filePath string, processor func(line string)) (err error) {
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

func ParseEpdTest(s string) *TestItem {
	var bmBegin = strings.Index(s, "bm")
	var bmEnd = strings.Index(s, ";")
	var fen = strings.TrimSpace(s[:bmBegin])
	var p = common.NewPositionFromFEN(fen)
	var sBestMoves = strings.Split(s[bmBegin:bmEnd], " ")[1:]
	var bestMoves []common.Move
	for _, sBestMove := range sBestMoves {
		var move = ParseEpdMove(p, sBestMove)
		if move == common.MoveEmpty {
			return nil
		}
		bestMoves = append(bestMoves, move)
	}
	if len(bestMoves) == 0 {
		return nil
	}
	return &TestItem{
		Content:   s,
		Position:  p,
		BestMoves: bestMoves,
	}
}

func ParseEpdMove(p *common.Position, s string) common.Move {
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
