package trainer

import (
	"encoding/binary"
	"io"
	"math"
	"os"
)

type Topology struct {
	Inputs        uint32
	Outputs       uint32
	HiddenNeurons []uint32
}

// Network is a neural network with 3 layers
type Network struct {
	Id       uint32
	Topology Topology
	Weights  []Matrix
	Biases   []Matrix
}

func (t *Topology) LayerSize() int {
	return len(t.HiddenNeurons) + 1
}

// Binary specification for the NNUE file:
// - All the data is stored in little-endian layout
// - All the matrices are written in column-major
// - The magic number/version consists of 4 bytes (int32):
//   - 66 (which is the ASCII code for B), uint8
//   - 90 (which is the ASCII code for Z), uint8
//   - 2 The major part of the current version number, uint8
//   - 0 The minor part of the current version number, uint8
//
// - 4 bytes (int32) to denote the network ID
// - 4 bytes (int32) to denote input size
// - 4 bytes (int32) to denote output size
// - 4 bytes (int32) number to represent the number of inputs
// - 4 bytes (int32) for the size of each layer
// - All weights for a layer, followed by all the biases of the same layer
// - Other layers follow just like the above point
func (n *Network) Save(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write headers
	buf := []byte{66, 90, 2, 0}
	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	// Write network Id
	binary.LittleEndian.PutUint32(buf, n.Id)
	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	// Write Topology
	buf = make([]byte, 3*4+4*len(n.Topology.HiddenNeurons))
	binary.LittleEndian.PutUint32(buf[0:], n.Topology.Inputs)
	binary.LittleEndian.PutUint32(buf[4:], n.Topology.Outputs)
	binary.LittleEndian.PutUint32(buf[8:], uint32(len(n.Topology.HiddenNeurons)))
	for i := 0; i < len(n.Topology.HiddenNeurons); i++ {
		binary.LittleEndian.PutUint32(buf[12+4*i:], n.Topology.HiddenNeurons[i])
	}
	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	var layerSize = n.Topology.LayerSize()
	for i := 0; i < layerSize; i++ {
		err = writeSlice(f, n.Weights[i].Data)
		if err != nil {
			return err
		}
		err = writeSlice(f, n.Biases[i].Data)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeSlice(f io.Writer, data []float64) error {
	buf := make([]byte, 4)
	for j := range data {
		binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(data[j])))
		_, err := f.Write(buf)
		if err != nil {
			return err
		}
	}
	return nil
}
