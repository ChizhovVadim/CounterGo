package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ChizhovVadim/CounterGo/internal/domain"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Data struct {
	fen    string
	target float64
}

type Sample struct {
	Target float32
	domain.TuneEntry
}

func LoadDataset(filepath string, e ITunableEvaluator,
	parser func(string) (Data, error)) ([]Sample, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []Sample

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()
		var data, err = parser(s)
		if err != nil {
			return nil, fmt.Errorf("parse fail %v %w", s, err)
		}

		pos, err := common.NewPositionFromFEN(data.fen)
		if err != nil {
			return nil, err
		}

		result = append(result, Sample{
			Target:    float32(data.target),
			TuneEntry: e.ComputeFeatures(&pos),
		})

		if config.datasetMaxSize != 0 && len(result) >= config.datasetMaxSize {
			log.Println("Limit dataset to prevent swap RAM")
			break
		}
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func mainParser(line string) (Data, error) {
	var fileds = strings.SplitN(line, ";", 2)
	if len(fileds) < 2 {
		return Data{}, fmt.Errorf("bad line")
	}

	var fen = fileds[0]
	var sResult = fileds[1]
	gameResult, err := strconv.ParseFloat(sResult, 64)
	if err != nil {
		return Data{}, err
	}

	return Data{
		fen:    fen,
		target: gameResult,
	}, nil
}

func zurichessParser(s string) (Data, error) {
	var index = strings.Index(s, "\"")
	if index < 0 {
		return Data{}, fmt.Errorf("bad separator")
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
		return Data{}, fmt.Errorf("bad game result")
	}

	return Data{
		fen:    fen,
		target: prob,
	}, nil
}
