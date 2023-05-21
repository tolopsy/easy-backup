package archiver

import (
	"io"
	"os"
	"path/filepath"
)

type localArchiver struct{}

func (z *localArchiver) Archive(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0777); err != nil {
		return err
	}

	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	zipReader, err := zipFile(source)
	if err != nil {
		return err
	}
	_, err = io.Copy(destinationFile, zipReader)

	return err
}

var LocalZipper Archiver = (*localArchiver)(nil)
