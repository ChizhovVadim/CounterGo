package train

import (
	"encoding/binary"
	"io"
	"math"
	"os"

	"github.com/ChizhovVadim/CounterGo/internal/ml"
)

type Topology struct {
	Inputs        uint32
	Outputs       uint32
	HiddenNeurons []uint32
}

func NewTopology(inputs, outputs uint32, hiddenNeurons []uint32) Topology {
	return Topology{
		Inputs:        inputs,
		Outputs:       outputs,
		HiddenNeurons: hiddenNeurons,
	}
}

// Network is a neural network with 3 layers
type Network struct {
	Id       uint32
	Topology Topology
	Weights  []ml.Matrix
	Biases   []ml.Matrix
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

// load a neural network from file
func LoadNetwork(path string) Network {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Read headers
	buf := make([]byte, 4)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		panic(err)
	}
	if buf[0] != 66 || buf[1] != 90 {
		panic("Magic word does not match expected, exiting")
	}

	if buf[2] != 2 || buf[3] != 0 {
		panic("Network binary format is not supported")
	}

	_, err = io.ReadFull(f, buf)
	if err != nil {
		panic(err)
	}
	id := binary.LittleEndian.Uint32(buf)

	// Read Topology Header
	buf = make([]byte, 12)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		panic(err)
	}
	inputs := binary.LittleEndian.Uint32(buf[:4])
	outputs := binary.LittleEndian.Uint32(buf[4:8])
	layers := binary.LittleEndian.Uint32(buf[8:])

	buf = make([]byte, 4*layers)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		panic(err)
	}
	neurons := make([]uint32, layers)
	for i := uint32(0); i < layers; i++ {
		neurons[i] = binary.LittleEndian.Uint32(buf[i*4 : (i+1)*4])
	}

	topology := NewTopology(inputs, outputs, neurons)

	net := Network{
		Topology: topology,
		Id:       id,
	}

	net.Weights = make([]ml.Matrix, len(topology.HiddenNeurons)+1)
	net.Biases = make([]ml.Matrix, len(topology.HiddenNeurons)+1)

	buf = make([]byte, 4)
	inputSize := int(topology.Inputs)

	for i := 0; i < len(topology.HiddenNeurons)+1; i++ {
		var outputSize int
		if i == len(neurons) {
			outputSize = int(outputs)
		} else {
			outputSize = int(neurons[i])
		}
		data := make([]float64, outputSize*inputSize)
		for j := 0; j < len(data); j++ {
			_, err := io.ReadFull(f, buf)
			if err != nil {
				panic(err)
			}
			data[j] = float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))
		}
		net.Weights[i] = ml.Matrix{Data: data, Rows: outputSize, Cols: inputSize}
		inputSize = outputSize

		data = make([]float64, outputSize)
		for j := 0; j < len(data); j++ {
			_, err := io.ReadFull(f, buf)
			if err != nil {
				panic(err)
			}
			data[j] = float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))
		}
		net.Biases[i] = ml.Matrix{Data: data, Rows: outputSize, Cols: 1}
	}
	return net
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
