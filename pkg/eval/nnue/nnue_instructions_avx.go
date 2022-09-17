//go:build avx
// +build avx

package eval

import "unsafe"

//go:noescape
func _update_hidden(previous_outputs, update_indices, update_coeffs, update_size, weights, outputs, outputs_len unsafe.Pointer)

//go:noescape
func _quick_feed(hidden_outputs, hidden_outputs_len, weights, weights_len, res unsafe.Pointer)

func (e *EvaluationService) UpdateHidden() {
	e.currentHidden++

	p1 := unsafe.Pointer(&e.hiddenOutputs[e.currentHidden-1][0])
	p2 := unsafe.Pointer(&e.updates.Indices[0])
	p3 := unsafe.Pointer(&e.updates.Coeffs[0])
	p4 := unsafe.Pointer(uintptr(e.updates.Size))
	p5 := unsafe.Pointer(&e.HiddenWeights[0])
	p6 := unsafe.Pointer(&e.hiddenOutputs[e.currentHidden][0])
	p7 := unsafe.Pointer(uintptr(HiddenSize))

	_update_hidden(p1, p2, p3, p4, p5, p6, p7)
}

func (e *EvaluationService) QuickFeed() float32 {
	p1 := unsafe.Pointer(&e.hiddenOutputs[e.currentHidden][0])
	p2 := unsafe.Pointer(uintptr(HiddenSize))
	p3 := unsafe.Pointer(&e.OutputWeights[0])
	p4 := unsafe.Pointer(uintptr(HiddenSize))
	var res float32

	_quick_feed(p1, p2, p3, p4, unsafe.Pointer(&res))
	return res + e.OutputBias
}
