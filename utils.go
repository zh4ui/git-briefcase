package main

import (
	"os"
	"path/filepath"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func statAbsDir(dir string) error {
	if !filepath.IsAbs(dir) {
		return Error("not an absolute path")
	}
	return statDir(dir)
}

func statDir(dir string) error {
	if fileInfo, err := os.Stat(dir); err != nil {
		return err
	} else {
		if !fileInfo.IsDir() {
			return Error("not a directory")
		}
	}
	return nil
}
