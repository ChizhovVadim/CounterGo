package eval

import (
	"encoding/binary"
	"io"
	"math"
)

func LoadWeights(f io.Reader) (*Weights, error) {
	// Read headers
	buf := make([]byte, 4)

	_, err := io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}

	// Read Topology Header
	buf = make([]byte, 12)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}

	buf = make([]byte, 4)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}

	var w = &Weights{}

	for i := 0; i < InputSize; i++ {
		for j := 0; j < HiddenSize; j++ {
			_, err := io.ReadFull(f, buf)
			if err != nil {
				return nil, err
			}
			w.HiddenWeights[i*HiddenSize+j] = math.Float32frombits(binary.LittleEndian.Uint32(buf))
		}
	}

	for i := 0; i < HiddenSize; i++ {
		_, err := io.ReadFull(f, buf)
		if err != nil {
			return nil, err
		}
		w.HiddenBiases[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf))
	}

	for i := 0; i < HiddenSize; i++ {
		_, err := io.ReadFull(f, buf)
		if err != nil {
			return nil, err
		}
		w.OutputWeights[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf))
	}

	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	w.OutputBias = math.Float32frombits(binary.LittleEndian.Uint32(buf))

	return w, nil
}
