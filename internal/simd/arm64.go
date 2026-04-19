//go:build arm64

package simd

//go:noescape
func addScaledNEON(dst []float32, alpha float32, s []float32)

//go:noescape
func reluNEON(dst, src []float32)

//go:noescape
func dotProductNEON(a, b []float32) float32

func Relu(dst, src []float32) {
	reluNEON(dst, src)
}

func DotProduct(a, b []float32) float32 {
	return dotProductNEON(a, b)
}

func AddScaled(dst []float32, alpha float32, s []float32) {
	addScaledNEON(dst, alpha, s)
}
