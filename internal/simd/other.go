//go:build !amd64 && !arm64

package simd

func Relu(dst, src []float32) {
	reluGo(dst, src)
}

func DotProduct(a, b []float32) float32 {
	return dotProductGo(a, b)
}

func AddScaled(dst []float32, alpha float32, s []float32) {
	addScaledGo(dst, alpha, s)
}

func reluGo(dst, src []float32) {
	for i := range dst {
		dst[i] = max(src[i], 0)
	}
}

func dotProductGo(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func addScaledGo(dst []float32, alpha float32, s []float32) {
	for i := range dst {
		dst[i] += alpha * s[i]
	}
}
