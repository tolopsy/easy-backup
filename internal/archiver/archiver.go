package archiver

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

type Archiver interface {
	GetBackupFileFormat() string
	Archive(source, destination string) error
}

type zipper struct{}

func (z *zipper) GetBackupFileFormat() string {
	return "%d.zip"
}

func (z *zipper) Archive(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0777); err != nil {
		return err
	}
	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	writer := zip.NewWriter(destinationFile)
	defer writer.Close()

	relevantPathIndexStarter := len(filepath.Dir(source)) + 1
	return filepath.Walk(source, func(path string, content os.FileInfo, err error) error {
		if content.IsDir() {
			return nil		// skip
		}
		if err != nil {
			return err
		}
		fileToBackup, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToBackup.Close()

		backupFile, err := writer.Create(path[relevantPathIndexStarter:])
		if err != nil {
			return err
		}

		_, err = io.Copy(backupFile, fileToBackup)
		if err != nil {
			return err
		}
		return err
	})
}

var ZIP Archiver = (*zipper)(nil)
