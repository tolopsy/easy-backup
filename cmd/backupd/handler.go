package main

import (
	"easy_backup/internal/archiver"
	"easy_backup/utils"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	mu            sync.Mutex
	wg            sync.WaitGroup
	Paths         map[string]string
	archiver.Archiver
	Destination string
}

func (handler *Handler) startRunWorker(id int, p <-chan utils.Path, results chan<- utils.Path, resultCounter *int, errChan chan<- error) {
	for path := range p {
		fmt.Printf("Worker %v processing path %s\n", id, path)
		newHash, err := utils.HashFile(path.Path)
		if err != nil {
			errChan <- err
			handler.wg.Done()
			continue
		}

		// do not run backup if hash hasn't changed.
		if newHash == path.Hash {
			handler.wg.Done()
			continue
		}

		if err = handler.backup(path.Path); err != nil {
			errChan <- err
			handler.wg.Done()
			continue
		}
		path.Hash = newHash
		results <- path

		handler.mu.Lock()
		*resultCounter++
		handler.mu.Unlock()
		
		handler.wg.Done()
	}
}

func (handler *Handler) Run() (int, []error) {
	var counter int
	errorList := make([]error, 0)
	numRunWorkers := 2
	pathChan := make(chan utils.Path, len(handler.Paths))
	resultChan := make(chan utils.Path, len(handler.Paths))

	// channel to pass error buffered by number of run workers
	errChan := make(chan error, numRunWorkers)

	// start the workers
	for id := 1; id <= numRunWorkers; id++ {
		go handler.startRunWorker(id, pathChan, resultChan, &counter, errChan)
	}

	handler.wg.Add(len(handler.Paths))
	for path, lastHash := range handler.Paths {
		pathChan <- utils.Path{Path: path, Hash: lastHash}
	}

	go func() {
		handler.wg.Wait()
		close(pathChan)
		close(errChan)
	}()
	
	// collect errors
	for errValue := range errChan {
		errorList = append(errorList, errValue)
	}

	// update hash for processed paths
	for resultCount := 1; resultCount <= counter; resultCount++ {
		path := <- resultChan
		handler.Paths[path.Path] = path.Hash
	}
	close(resultChan)

	return counter, errorList
}

func (handler *Handler) backup(path string) error {
	dirName := filepath.Base(path)
	fileName := fmt.Sprintf(archiver.GetBackupFileFormat(), time.Now().UnixNano())
	locationElem := []string{handler.Destination, dirName, fileName}

	destination := filepath.Join(locationElem...)
	if _, ok := handler.Archiver.(*archiver.S3Archiver); ok {
		destination = strings.Join(locationElem, "/")
	}

	return handler.Archive(path, destination)
}
