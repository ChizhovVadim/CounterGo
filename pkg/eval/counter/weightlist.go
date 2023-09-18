package eval

import "github.com/ChizhovVadim/CounterGo/internal/domain"

type WeightList struct {
	weights []Score
	values  []int
	tuning  bool
}

func (wl *WeightList) init(size int) {
	wl.weights = make([]Score, size)
	wl.values = make([]int, size)
}

func (wl *WeightList) Value(index, n int) Score {
	if wl.tuning {
		wl.values[index] += n
	}
	return wl.weights[index] * Score(n)
}

func (wl *WeightList) InitWeights(w []int) {
	if 2*len(wl.weights) != len(w) {
		return
	}
	for i := range wl.weights {
		wl.weights[i] = S(w[2*i], w[2*i+1])
	}
}

func (wl *WeightList) Features() []domain.FeatureInfo {
	var size int
	for _, v := range wl.values {
		if v != 0 {
			size++
		}
	}
	var features = make([]domain.FeatureInfo, 0, size)
	for i, v := range wl.values {
		if v != 0 {
			features = append(features, domain.FeatureInfo{Index: int16(i), Value: int16(v)})
		}
	}
	return features
}
