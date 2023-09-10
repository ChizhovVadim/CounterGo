//go:build embed
// +build embed

package eval

import (
	"embed"
)

//go:embed n-30-5094.nn
var content embed.FS

func loadEmbedWeights() (*Weights, error) {
	const name = "n-30-5094.nn"
	var f, err = content.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadWeights(f)
}
