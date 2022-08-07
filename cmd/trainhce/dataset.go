package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ChizhovVadim/CounterGo/internal/domain"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Sample struct {
	Target float32
	domain.TuneEntry
}

func LoadDataset(filepath string, e ITunableEvaluator,
	parser func(string, ITunableEvaluator) (Sample, error)) ([]Sample, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []Sample

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()
		var sample, err = parser(s, e)
		if err != nil {
			return nil, fmt.Errorf("parse fail %v %w", s, err)
		}
		result = append(result, sample)
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseValidationSample(s string, e ITunableEvaluator) (Sample, error) {
	var sample Sample

	var index = strings.Index(s, "\"")
	if index < 0 {
		return Sample{}, fmt.Errorf("bad separator")
	}

	var fen = s[:index]
	var strScore = s[index+1:]

	var pos, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return Sample{}, err
	}
	sample.TuneEntry = e.ComputeFeatures(&pos)

	var prob float32
	if strings.HasPrefix(strScore, "1/2-1/2") {
		prob = 0.5
	} else if strings.HasPrefix(strScore, "1-0") {
		prob = 1.0
	} else if strings.HasPrefix(strScore, "0-1") {
		prob = 0.0
	} else {
		return Sample{}, fmt.Errorf("bad game result")
	}
	sample.Target = prob

	return sample, nil
}

func parseTrainingSample(line string, e ITunableEvaluator) (Sample, error) {
	var sample Sample

	var fileds = strings.SplitN(line, ";", 3)
	if len(fileds) < 3 {
		return Sample{}, fmt.Errorf("Bad line")
	}

	var fen = fileds[0]
	var pos, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return Sample{}, err
	}
	sample.TuneEntry = e.ComputeFeatures(&pos)

	var sScore = fileds[1]
	score, err := strconv.Atoi(sScore)
	if err != nil {
		return Sample{}, err
	}

	var sResult = fileds[2]
	gameResult, err := strconv.ParseFloat(sResult, 64)
	if err != nil {
		return Sample{}, err
	}

	const W = 0.75
	var prob = W*Sigmoid(float64(score)) + (1-W)*gameResult
	sample.Target = float32(prob)

	return sample, nil
}
