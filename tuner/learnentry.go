package tuner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type learnEntry struct {
	fen   string
	score float64
}

func readLearn(filename string, learnEntries chan<- learnEntry) error {
	file, err := os.Open(filename)
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

func parseLearnEntry(s string) (learnEntry, error) {
	var index = strings.Index(s, "\"")
	if index < 0 {
		return learnEntry{}, fmt.Errorf("parseLearnEntry failed %v", s)
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
		return learnEntry{}, fmt.Errorf("parseLearnEntry failed %v", s)
	}
	return learnEntry{fen, score}, nil
}
