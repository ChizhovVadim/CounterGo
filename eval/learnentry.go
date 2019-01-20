package eval

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type LearnEntry struct {
	Fen   string
	Score float64
}

const LearnFilePath = "/home/vadim/chess/tuner/quiet-labeled.epd"

func ReadLearn(filePath string, learnEntries chan<- LearnEntry) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var entry, err = parseLearnEntry(line)
		if err != nil {
			return err
		}
		learnEntries <- entry
	}
	return scanner.Err()
}

func parseLearnEntry(s string) (LearnEntry, error) {
	var index = strings.Index(s, "\"")
	if index < 0 {
		return LearnEntry{}, fmt.Errorf("parseLearnEntry failed %v", s)
	}
	var fen = s[:index]
	var score float64
	var strScore = s[index+1:]
	if strings.HasPrefix(strScore, "1/2-1/2") {
		score = 0.5
	} else if strings.HasPrefix(strScore, "1-0") {
		score = 1.0
	} else if strings.HasPrefix(strScore, "0-1") {
		score = 0.0
	} else {
		return LearnEntry{}, fmt.Errorf("parseLearnEntry failed %v", s)
	}
	return LearnEntry{fen, score}, nil
}
