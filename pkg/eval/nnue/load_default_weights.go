//go:build !embed
// +build !embed

package eval

import (
	"log"
)

func loadDefaultWeights() (*Weights, error) {
	var path = mapPath("./n-30-5094.nn")
	w, err := loadFileWeights(path)
	if err == nil {
		log.Println("loaded nnue weights", "path", path)
		return w, nil
	}
	path = mapPath("~/chess/n-30-5094.nn")
	w, err = loadFileWeights(path)
	if err == nil {
		log.Println("loaded nnue weights", "path", path)
		return w, nil
	}
	return nil, err
}
