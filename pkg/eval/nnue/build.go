package eval

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

func loadWeightsСached(loadFunc func() (*Weights, error)) func() (*Weights, error) {
	var once sync.Once
	var weights *Weights
	var err error
	return func() (*Weights, error) {
		once.Do(func() {
			weights, err = loadFunc()
		})
		return weights, err
	}
}

var loadDefaultWeightsСached = loadWeightsСached(loadDefaultWeights)

// TODO return err
func NewDefaultEvaluationService() *EvaluationService {
	var weights, err = loadDefaultWeightsСached()
	if err != nil {
		panic(err)
	}
	return NewEvaluationService(weights)
}

func loadFileWeights(path string) (*Weights, error) {
	var f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadWeights(f)
}

func mapPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		curUser, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(curUser.HomeDir, strings.TrimPrefix(path, "~/"))
	}
	if strings.HasPrefix(path, "./") {
		var exePath, err = os.Executable()
		if err != nil {
			return path
		}
		return filepath.Join(filepath.Dir(exePath), strings.TrimPrefix(path, "./"))
	}
	return path
}
