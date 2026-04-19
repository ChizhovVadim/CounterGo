//go:build embed

package evalnn

import "embed"

//go:embed n-30-5268.nn
var content embed.FS

func Load() (*Weights, error) {
	const name = "n-30-5268.nn"
	var f, err = content.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadWeights(f, true)
}
