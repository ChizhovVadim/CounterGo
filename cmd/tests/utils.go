package main

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

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
