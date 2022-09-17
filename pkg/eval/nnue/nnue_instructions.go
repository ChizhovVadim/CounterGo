//go:build !avx
// +build !avx

package eval

func (e *EvaluationService) QuickFeed() float32 {
	var output float32
	for i, x := range e.hiddenOutputs[e.currentHidden][:] {
		if x > 0 {
			output += x * e.OutputWeights[i]
		}
	}
	return output + e.OutputBias
}

func (e *EvaluationService) UpdateHidden() {
	e.currentHidden++
	hiddenOutputs := e.hiddenOutputs[e.currentHidden][:]
	copy(hiddenOutputs, e.hiddenOutputs[e.currentHidden-1][:])

	for i := 0; i < e.updates.Size; i++ {
		var index = int(e.updates.Indices[i]) * HiddenSize
		if e.updates.Coeffs[i] == Add {
			for j := range hiddenOutputs {
				hiddenOutputs[j] += e.HiddenWeights[index+j]
			}
		} else {
			for j := range hiddenOutputs {
				hiddenOutputs[j] -= e.HiddenWeights[index+j]
			}
		}
	}
}
