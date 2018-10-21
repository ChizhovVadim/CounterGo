package learn

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/ChizhovVadim/CounterGo/common"
	"github.com/ChizhovVadim/CounterGo/engine"
)

type tuningEntry struct {
	p     common.Position
	score float32
}

func RunTuning() {
	log.Println("Tune started.")
	var entries, err = readTuningFile("/home/vadim/chess/tuner/quiet-labeled.epd")
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Loaded %v entries.\n", len(entries))
	var params, errorFunc = newErrorFunc(entries)
	tuneEvalService(params, errorFunc)
	log.Println("Tune finished.")
}

func ShowError() {
	var entries, err = readTuningFile("/home/vadim/chess/tuner/quiet-labeled.epd")
	if err != nil {
		log.Println(err)
		return
	}
	var _, errorFunc = newErrorFunc(entries)
	var es = engine.NewEvaluationService()
	var e = errorFunc(es.DefaultParams())
	log.Println("Error:", e)
}

func readTuningFile(filePath string) ([]tuningEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var result []tuningEntry
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		var entry, err = parseTuningEntry(line)
		if err != nil {
			log.Println(err)
			continue
		}
		result = append(result, entry)
	}
	return result, nil
}

func parseTuningEntry(s string) (tuningEntry, error) {
	var index = strings.Index(s, "\"")
	if index < 0 {
		return tuningEntry{}, fmt.Errorf("parseTuningEntry failed %v", s)
	}
	pos, err := common.NewPositionFromFEN(s[:index])
	if err != nil {
		return tuningEntry{}, fmt.Errorf("parseTuningEntry failed %v", s)
	}
	var score float64
	var strScore = s[index+1:]
	if strings.HasPrefix(strScore, "1/2-1/2") {
		score = 0.5
	} else if strings.HasPrefix(strScore, "1-0") {
		score = 1.0
	} else if strings.HasPrefix(strScore, "0-1") {
		score = 0.0
	} else {
		return tuningEntry{}, fmt.Errorf("parseTuningEntry failed %v", s)
	}
	return tuningEntry{p: pos, score: float32(score)}, nil
}

//https://www.chessprogramming.org/Texel%27s_Tuning_Method
func tuneEvalService(params []int, errorFunc func([]int) float64) {
	var bestE = errorFunc(params)
	for iter := 0; iter < 50; iter++ {
		log.Printf("Iteration: %v Error: %.6f Params: %#v\n",
			iter, bestE, params)
		var improved = false
		for featureIndex := range params {
			var oldValue = params[featureIndex]
			var bestValue = oldValue
			for step := 1; step <= 64; step *= 2 {
				params[featureIndex] = bestValue + step
				var newE = errorFunc(params)
				if newE < bestE {
					bestValue = params[featureIndex]
					bestE = newE
					improved = true
				} else {
					params[featureIndex] = bestValue
					if newE > bestE {
						break
					}
				}
			}
			if oldValue == bestValue {
				for step := 1; step <= 64; step *= 2 {
					params[featureIndex] = bestValue - step
					var newE = errorFunc(params)
					if newE < bestE {
						bestValue = params[featureIndex]
						bestE = newE
						improved = true
					} else {
						params[featureIndex] = bestValue
						if newE > bestE {
							break
						}
					}
				}
			}
		}
		if !improved {
			break
		}
	}
	fmt.Printf("// Error: %.6f\n", bestE)
	fmt.Printf("return %#v\n", params)
}

func newErrorFunc(entries []tuningEntry) (params []int, errorFunc func([]int) float64) {
	var numCPUs = runtime.NumCPU()
	var evalServices = make([]*engine.EvaluationService, numCPUs)
	for i := range evalServices {
		evalServices[i] = engine.NewEvaluationService()
	}
	params = evalServices[0].Init(nil)
	errorFunc = func(params []int) float64 {
		var reg float64
		var sums = make([]float64, numCPUs)
		var wg = &sync.WaitGroup{}
		for i := 0; i < numCPUs; i++ {
			wg.Add(1)
			go func(thread int) {
				defer wg.Done()
				var e = evalServices[thread]
				e.Init(params)
				if thread == 0 {
					const lambda = 1e-6
					reg = lambda * float64(e.Regularization())
				}
				var localSum = 0.0
				//TODO sync.atomic Add(&i, 1)
				for i := thread; i < len(entries); i += numCPUs {
					var entry = &entries[i]
					var score = e.Evaluate(&entry.p)
					if !entry.p.WhiteMove {
						score = -score
					}
					var expected = mixProbability(
						float64(entry.score),
						sigmoid(evalMaterial(&entry.p)),
						0.25,
						0.5)
					var diff = sigmoid(score) - expected
					localSum += diff * diff
				}
				sums[thread] = localSum
			}(i)
		}
		wg.Wait()
		var sum = 0.0
		for _, item := range sums {
			sum += item
		}
		return reg + sum/float64(len(entries))
	}
	return
}

func sigmoid(s int) float64 {
	return 1.0 / (1.0 + math.Exp(-float64(s)/150))
}

func evalMaterial(p *common.Position) int {
	return 100*(common.PopCount(p.Pawns&p.White)-common.PopCount(p.Pawns&p.Black)) +
		325*(common.PopCount(p.Knights&p.White)-common.PopCount(p.Knights&p.Black)) +
		325*(common.PopCount(p.Bishops&p.White)-common.PopCount(p.Bishops&p.Black)) +
		500*(common.PopCount(p.Rooks&p.White)-common.PopCount(p.Rooks&p.Black)) +
		1000*(common.PopCount(p.Queens&p.White)-common.PopCount(p.Queens&p.Black))
}

func limitValue(v, lower, upper float64) float64 {
	if v < lower {
		return lower
	}
	if v > upper {
		return upper
	}
	return v
}

func mixProbability(mainProb, materialProb, materialFree, materialWeight float64) float64 {
	return mainProb + materialWeight*(limitValue(mainProb, materialProb-materialFree, materialProb+materialFree)-mainProb)
}
