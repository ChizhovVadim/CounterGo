//go:build embed
// +build embed

package eval

import (
	"embed"
	"log"
)

//go:embed n-30-5268.nn
var content embed.FS

func loadDefaultWeights() (*Weights, error) {
	const name = "n-30-5268.nn"
	var f, err = content.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	w, err := LoadWeights(f)
	if err != nil {
		return nil, err
	}
	log.Println("loaded embed nnue weights")
	return w, nil
}
