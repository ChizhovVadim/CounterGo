package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func walkDataset(
	datasetPath string,
	onItem func(pos *common.Position, gameResult float64) error,
) error {

	file, err := os.Open(datasetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()

		var index = strings.Index(s, "\"")
		if index < 0 {
			return fmt.Errorf("zurichessParser failed %v", s)
		}

		var fen = s[:index]
		pos, err := common.NewPositionFromFEN(fen)
		if err != nil {
			return err
		}

		var strScore = s[index+1:]
		var prob float64
		if strings.HasPrefix(strScore, "1/2-1/2") {
			prob = 0.5
		} else if strings.HasPrefix(strScore, "1-0") {
			prob = 1.0
		} else if strings.HasPrefix(strScore, "0-1") {
			prob = 0.0
		} else {
			return fmt.Errorf("zurichessParser failed %v", s)
		}

		err = onItem(&pos, prob)
		if err != nil {
			return err
		}
	}

	return nil
}
