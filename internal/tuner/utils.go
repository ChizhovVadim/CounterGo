package tuner

var (
	LearningRate = 0.01
)

func ValidationCost(output, target float64) float64 {
	var x = output - target
	return x * x
}

func CalculateCostGradient(output, target float64) float64 {
	return 2.0 * (output - target)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
