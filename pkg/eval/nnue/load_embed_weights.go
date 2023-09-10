//go:build !embed
// +build !embed

package eval

import "fmt"

func loadEmbedWeights() (*Weights, error) {
	return nil, fmt.Errorf("no embed weights")
}
