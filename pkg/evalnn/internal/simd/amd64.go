//go:build amd64

package simd

//go:noescape
func dotProductAVX(a, b []float32) float32

//go:noescape
func addScaledAVX(dst []float32, alpha float32, s []float32)

//go:noescape
func reluAVX(dst, src []float32)

//go:noescape
func dotProductAVX512(a, b []float32) float32

//go:noescape
func addScaledAVX512(dst []float32, alpha float32, s []float32)

func Relu(dst, src []float32) {
	reluAVX(dst, src)
}

func DotProduct(a, b []float32) float32 {
	return dotProductAVX(a, b)
}

func AddScaled(dst []float32, alpha float32, s []float32) {
	addScaledAVX(dst, alpha, s)
}
