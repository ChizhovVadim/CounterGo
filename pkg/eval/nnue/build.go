package eval

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

var once sync.Once
var weights *Weights

// TODO return err
func NewDefaultEvaluationService() *EvaluationService {
	once.Do(func() {
		var w, err = loadEmbedWeights()
		if err == nil {
			weights = w
			log.Println("loaded embed nnue weights")
			return
		}
		var path = mapPath("./n-30-5094.nn")
		w, err = loadFileWeights(path)
		if err == nil {
			weights = w
			log.Println("loaded nnue weights", "path", path)
			return
		}
		path = mapPath("~/chess/n-30-5094.nn")
		w, err = loadFileWeights(path)
		if err == nil {
			weights = w
			log.Println("loaded nnue weights", "path", path)
			return
		}
		panic(err)
	})
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
