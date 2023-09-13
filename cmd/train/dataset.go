package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Sample struct {
	Input  []int16
	Target float32
}

func LoadZurichessDataset(filepath string) ([]Sample, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []Sample

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var sample, err = zurichessParser(line)
		if err != nil {
			return nil, err
		}
		result = append(result, sample)
	}

	return result, nil
}

func zurichessParser(s string) (Sample, error) {
	var index = strings.Index(s, "\"")
	if index < 0 {
		return Sample{}, fmt.Errorf("zurichessParser failed %v", s)
	}

	var fen = s[:index]
	input, err := FromFen(fen)
	if err != nil {
		return Sample{}, err
	}

	var strScore = s[index+1:]
	var prob float32
	if strings.HasPrefix(strScore, "1/2-1/2") {
		prob = 0.5
	} else if strings.HasPrefix(strScore, "1-0") {
		prob = 1.0
	} else if strings.HasPrefix(strScore, "0-1") {
		prob = 0.0
	} else {
		return Sample{}, fmt.Errorf("zurichessParser failed %v", s)
	}

	return Sample{input, prob}, nil
}

func LoadDataset(filepath string) ([]Sample, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []Sample

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()

		var fileds = strings.SplitN(line, ";", 2)
		if len(fileds) < 2 {
			return nil, fmt.Errorf("bad line %s", line)
		}

		var sFen = fileds[0]
		pos, err := common.NewPositionFromFEN(sFen)
		if err != nil {
			return nil, fmt.Errorf("bad fen %s", line)
		}
		var sTarget = fileds[1]
		target, err := strconv.ParseFloat(sTarget, 64)
		if err != nil {
			return nil, fmt.Errorf("bad target %s", line)
		}
		input, err := FromFen(sFen)
		if err != nil {
			return nil, err
		}
		result = append(result, Sample{
			Input:  input,
			Target: float32(target),
		})

		var mirrorPos = common.MirrorPosition(&pos)
		mirrorInput, err := FromFen(mirrorPos.String())
		if err != nil {
			return nil, err
		}
		result = append(result, Sample{
			Input:  mirrorInput,
			Target: 1 - float32(target),
		})

		if config.datasetMaxSize != 0 && len(result) >= config.datasetMaxSize {
			log.Println("Limit dataset to prevent swap RAM")
			break
		}
	}

	return result, nil
}
