package quality

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type IEvaluator interface {
	EvaluateProb(pos *common.Position) float64
}

func RunQuality(evaluator IEvaluator, validationPath string) error {
	file, err := os.Open(validationPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var sum, sumSq float64
	var count int

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()
		var entry, err = parseEntry(s)
		if err != nil {
			return err
		}

		pos, err := common.NewPositionFromFEN(entry.fen)
		if err != nil {
			return err
		}

		var eval = evaluator.EvaluateProb(&pos)
		var x = eval - entry.target
		sum += math.Abs(x)
		sumSq += x * x
		count += 1
	}

	//var absCost = sum / float64(count)
	var mseCost = sumSq / float64(count)
	//log.Printf("abs cost: %f", absCost)
	log.Printf("mse cost: %f", mseCost)
	return nil
}

type Entry struct {
	fen    string
	target float64
}

func parseEntry(s string) (Entry, error) {
	var index = strings.Index(s, "\"")
	if index < 0 {
		return Entry{}, fmt.Errorf("zurichessParser failed %v", s)
	}

	var fen = s[:index]
	var strScore = s[index+1:]

	var gameResult float64
	if strings.HasPrefix(strScore, "1/2-1/2") {
		gameResult = 0.5
	} else if strings.HasPrefix(strScore, "1-0") {
		gameResult = 1.0
	} else if strings.HasPrefix(strScore, "0-1") {
		gameResult = 0.0
	} else {
		return Entry{}, fmt.Errorf("zurichessParser failed %v", s)
	}
	return Entry{
		fen:    fen,
		target: gameResult,
	}, nil
}
