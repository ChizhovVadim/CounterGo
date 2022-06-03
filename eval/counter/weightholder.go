package eval

type weightHolder struct {
	weights []int
	index   int
}

func (wh *weightHolder) withDefault(v int) *weightHolder {
	if wh.index >= len(wh.weights) {
		wh.weights = append(wh.weights, v)
	}
	return wh
}

func (wh *weightHolder) next() int {
	if wh.index >= len(wh.weights) {
		wh.weights = append(wh.weights, 0)
	}
	var value = wh.weights[wh.index]
	wh.index++
	return value
}

func (wh *weightHolder) nextScore() Score {
	return Score{wh.next(), wh.next()}
}

func (wh *weightHolder) initInts(data []int) {
	for i := range data {
		data[i] = wh.next()
	}
}

func (wh *weightHolder) initScores(data []Score) {
	for i := range data {
		data[i] = wh.nextScore()
	}
}
