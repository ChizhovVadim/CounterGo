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

func LoadDataset(filepath string,
	sampleParser func(line string) (Sample, error)) ([]Sample, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []Sample

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var sample, err = sampleParser(line)
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

type Data struct {
	fen    string
	score  int
	result float64
}

func parseData(line string) (Data, error) {
	var fileds = strings.SplitN(line, ";", 3)
	if len(fileds) < 3 {
		return Data{}, fmt.Errorf("bad line %s", line)
	}

	var sFen = fileds[0]

	var sScore = fileds[1]
	score, err := strconv.Atoi(sScore)
	if err != nil {
		return Data{}, err
	}

	var sResult = fileds[2]
	result, err := strconv.ParseFloat(sResult, 64)
	if err != nil {
		return Data{}, err
	}

	return Data{
		fen:    sFen,
		score:  score,
		result: result,
	}, nil
}

var sigmoid = &Sigmoid{SigmoidScale: SigmoidScale}

func dataToSample(data Data) (Sample, error) {
	input, err := FromFen(data.fen)
	if err != nil {
		return Sample{}, err
	}
	var W = config.searchWeight
	var prob = W*sigmoid.Sigma(float64(data.score)) + (1-W)*data.result
	return Sample{Input: input, Target: float32(prob)}, nil
}

func LoadDataset2(filepath string) ([]Sample, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []Sample

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var data, err = parseData(line)
		if err != nil {
			return nil, err
		}
		sample, err := dataToSample(data)
		if err != nil {
			return nil, err
		}
		sample2, err := dataToSample(mirrorData(data))
		if err != nil {
			return nil, err
		}
		result = append(result, sample, sample2)

		if config.datasetMaxSize != 0 && len(result) >= config.datasetMaxSize {
			log.Println("Limit dataset to prevent swap RAM")
			break
		}
	}

	return result, nil
}

func mirrorData(data Data) Data {
	var pos, err = common.NewPositionFromFEN(data.fen)
	if err != nil {
		panic(err)
	}
	var mpos = common.MirrorPosition(&pos)
	return Data{fen: mpos.String(), score: -data.score, result: 1 - data.result}
}
