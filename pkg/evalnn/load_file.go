//go:build !embed

package evalnn

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func Load() (*Weights, error) {
	var path = findFile("n-30-5268.nn")
	var f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadWeights(f, true)
}

func findFile(name string) string {
	return filepath.Join(mapPath("~/chess"), name)
}

func mapPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		curUser, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(curUser.HomeDir, strings.TrimPrefix(path, "~/"))
	}
	return path
}
