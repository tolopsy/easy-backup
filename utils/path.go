package utils

import (
	"fmt"
	"path/filepath"
)

type Path struct {
	Path string
	Hash string
}

func (p Path) String() string {
	return fmt.Sprintf("%s (%s)", p.Path, p.Hash)
}

func GetAbsPath(name string) string {
	absPath, _ := filepath.Abs(name)
	return absPath
}

var PathFileName = "Paths"
