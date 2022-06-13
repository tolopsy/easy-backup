package backup

import (
	"fmt"
	"path/filepath"
	"time"
)

type Handler struct {
	Paths map[string]string
	Archiver
	Destination string
}

func (handler *Handler) Run() (int, []error) {
	var counter int
	errorList := make([]error, 0, 2)
	for path, lastHash := range handler.Paths {
		newHash, err := HashFile(path)
		if err != nil {
			errorList = append(errorList, err)
			continue
		}

		if newHash != lastHash {
			err := handler.backup(path)
			if err != nil {
				errorList = append(errorList, err)
				continue
			}
			handler.Paths[path] = newHash
			counter++
		}
	}
	return counter, errorList
}

func (handler *Handler) backup(path string) error {
	dirName := filepath.Base(path)
	fileName := fmt.Sprintf(handler.GetBackupFileFormat(), time.Now().UnixNano())
	return handler.Archive(path, filepath.Join(handler.Destination, dirName, fileName))
}
