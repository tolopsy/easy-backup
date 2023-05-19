package backup

import (
	pathutils "easy_backup/path_utils"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

type Handler struct {
	mu            sync.Mutex
	wg            sync.WaitGroup
	Paths         map[string]string
	Archiver
	Destination string
}

func (handler *Handler) startRunWorker(id int, p <-chan pathutils.Path, results chan<- pathutils.Path, resultCounter *int, errChan chan<- error) {
	for path := range p {
		fmt.Printf("Worker %v processing path %s\n", id, path)
		newHash, err := HashFile(path.Path)
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
	pathChan := make(chan pathutils.Path, len(handler.Paths))
	resultChan := make(chan pathutils.Path, len(handler.Paths))

	// channel to pass error buffered by number of run workers
	errChan := make(chan error, numRunWorkers)

	// start the workers
	for id := 1; id <= numRunWorkers; id++ {
		go handler.startRunWorker(id, pathChan, resultChan, &counter, errChan)
	}

	handler.wg.Add(len(handler.Paths))
	for path, lastHash := range handler.Paths {
		pathChan <- pathutils.Path{Path: path, Hash: lastHash}
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
	fileName := fmt.Sprintf(handler.GetBackupFileFormat(), time.Now().UnixNano())
	return handler.Archive(path, filepath.Join(handler.Destination, dirName, fileName))
}
