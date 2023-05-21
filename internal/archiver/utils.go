package archiver

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func zipFile(source string) (io.Reader, error) {
	var buf bytes.Buffer

	writer := zip.NewWriter(&buf)
	defer writer.Close()

	relevantPathIndexStarter := len(filepath.Dir(source)) + 1
	err := filepath.Walk(source, func(path string, content os.FileInfo, err error) error {
		if content.IsDir() {
			return nil // skip
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

	return &buf, err
}

func GetBackupFileFormat() string {
	return "%d.zip"
}
