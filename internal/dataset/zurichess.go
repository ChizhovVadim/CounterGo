package dataset

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type DatasetItem struct {
	Fen    string
	Target float64
}

func LoadZurichessDataset(
	filePath string,
) ([]DatasetItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []DatasetItem

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()

		var index = strings.Index(s, "\"")
		if index < 0 {
			return nil, fmt.Errorf("zurichessParser failed %v", s)
		}

		var fen = s[:index]
		var strScore = s[index+1:]

		var prob float64
		if strings.HasPrefix(strScore, "1/2-1/2") {
			prob = 0.5
		} else if strings.HasPrefix(strScore, "1-0") {
			prob = 1.0
		} else if strings.HasPrefix(strScore, "0-1") {
			prob = 0.0
		} else {
			return nil, fmt.Errorf("zurichessParser failed %v", s)
		}

		result = append(result, DatasetItem{
			Fen:    fen,
			Target: prob,
		})
	}

	return result, nil
}
