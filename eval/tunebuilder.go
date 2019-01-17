package eval

import (
	"math"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/ChizhovVadim/CounterGo/common"
)

type tuneBuilder struct {
	evalService *EvaluationService
	numCPU      int
	samples     []tuneEntry
}

type tuneEntry struct {
	score     float64
	evalEntry evalEntry
}

func NewTuneBuilder() *tuneBuilder {
	var srv = &tuneBuilder{
		evalService: NewEvaluationService(),
		numCPU:      runtime.NumCPU(),
	}
	return srv
}

func (srv *tuneBuilder) AddSample(fen string, score float64) error {
	var p, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return err
	}
	var evalEntry = srv.evalService.computeEntry(&p)
	srv.samples = append(srv.samples, tuneEntry{score, evalEntry})
	return nil
}

func (srv *tuneBuilder) GetStartingWeights() []int {
	var result = make([]int, 2*fSize)
	copy(result, []int{100, 100, 325, 325, 325, 325, 500, 500, 1000, 1000})
	return result
}

func (srv *tuneBuilder) ComputeError(weights []int, lambda float64) float64 {
	var reg = l1(weights)
	var sums = make([]float64, srv.numCPU)
	var wg = &sync.WaitGroup{}
	var scaledWeights = scaleWeights(weights, srv.evalService.scale[:])
	var index = int32(0)
	for thread := 0; thread < srv.numCPU; thread++ {
		wg.Add(1)
		go func(thread int) {
			defer wg.Done()
			var localSum = 0.0
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(srv.samples) {
					break
				}
				var entry = &srv.samples[i]
				var score = entry.evalEntry.Evaluate(scaledWeights)
				var diff = float64(entry.score) - sigmoid(float64(score))
				localSum += diff * diff
			}
			sums[thread] = localSum
		}(thread)
	}
	wg.Wait()
	var sum = 0.0
	for _, item := range sums {
		sum += item
	}
	return sum/float64(len(srv.samples)) + lambda*reg
}

func scaleWeights(weights, scales []int) []int {
	var result = make([]int, len(weights))
	for i := range result {
		result[i] = weights[i] * evalScale / scales[i/2]
	}
	return result
}

func l1(weights []int) float64 {
	var reg = 0.0
	for i := 0; i < fSize; i++ {
		var x = float64(weights[2*i])
		var y = float64(weights[2*i+1])
		if math.Signbit(x) == math.Signbit(y) {
			reg += math.Max(math.Abs(x), math.Abs(y))
		} else {
			reg += math.Abs(x) + math.Abs(y)
		}
	}
	return reg
}

func sigmoid(s float64) float64 {
	return 1.0 / (1.0 + math.Exp(-s/150))
}
