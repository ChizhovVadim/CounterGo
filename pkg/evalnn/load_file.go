//go:build !embed

package evalnn

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func Load() (*Weights, error) {
	const name = "n-30-5268.nn"
	var path = findFile(name)
	if path == "" {
		return nil, fmt.Errorf("file not found %v", name)
	}
	var f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadWeights(f, true)
}

func findFile(name string) string {
	if wd, err := os.Getwd(); err == nil {
		var path = filepath.Join(wd, name)
		if pathExists(path) {
			return path
		}
	}
	if exePath, err := os.Executable(); err == nil {
		var path = filepath.Join(filepath.Dir(exePath), name)
		if pathExists(path) {
			return path
		}
	}
	if curUser, err := user.Current(); err == nil {
		var path = filepath.Join(curUser.HomeDir, "chess", name)
		if pathExists(path) {
			return path
		}
	}
	return ""
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // File exists
	}
	if errors.Is(err, os.ErrNotExist) {
		return false // File does not exist
	}
	return false // Schrodinger: Permission error or other issue
}
