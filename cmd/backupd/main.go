package main

import (
	backup "easy_backup"
	pathutils "easy_backup/path_utils"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/matryer/filedb"
)

type path = pathutils.Path

func main() {
	var fatalErr error
	defer func() {
		if fatalErr != nil {
			log.Println(fatalErr)
		}
	}()

	getAbsPath := pathutils.GetAbsPath
	workingDir, _ := os.Getwd()
	defaultDbPath := filepath.Join(filepath.Dir(workingDir), "backup", "data")
	var (
		interval = flag.Duration("interval", 10*time.Second, "Backup cycle: Interval between archive process")
		backupTo = flag.String("backup_dir", "backups", "Path to archive location")
		dbPath   = flag.String("db", defaultDbPath, "Filesystem DB storing paths of files to backup")
	)
	flag.Parse()

	backupHandler := &backup.Handler{
		Paths:       make(map[string]string),
		Archiver:    backup.ZIP,
		Destination: *backupTo,
	}

	db, err := filedb.Dial(getAbsPath(*dbPath))
	if err != nil {
		fatalErr = err
		return
	}
	defer db.Close()

	pathCollection, err := db.C(pathutils.PathFileName)
	if err != nil {
		fatalErr = err
		return
	}

	var path path
	err = pathCollection.ForEach(func(_ int, data []byte) bool {
		if err := json.Unmarshal(data, &path); err != nil {
			fatalErr = err
			return true
		}
		backupHandler.Paths[path.Path] = path.Hash
		return false
	})

	if err != nil {
		fatalErr = err
		return
	}
	if len(backupHandler.Paths) < 1 {
		fatalErr = errors.New("no path: add atleast one path using the backup tool")
		return
	}

	runBackupCycle(backupHandler, pathCollection)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-time.After(*interval):
			runBackupCycle(backupHandler, pathCollection)
		case <-signalChan:
			fmt.Println()
			log.Println("stopping")
			return
		}
	}
}

func runBackupCycle(handler *backup.Handler, collection *filedb.C) {
	log.Println("Running backup cycle")
	counter, errorList := handler.Run()
	if len(errorList) != 0 {
		log.Printf("Failed to backup - %+q\n\n", errorList)
	}

	if counter == 0 {
		log.Println("\tNo changes!")
		return
	}


	log.Printf("\t%d Directories backed up", counter)
	var path path

	// update hash of all paths
	collection.SelectEach(func(_ int, data []byte) (bool, []byte, bool) {
		if err := json.Unmarshal(data, &path); err != nil {
			log.Println("Failed to parse data (skipping...)", err)
			return true, data, false
		}
		path.Hash = handler.Paths[path.Path]
		newData, err := json.Marshal(path)
		if err != nil {
			log.Println("Failed to jsonify path struct (skipping...)", err)
			return true, data, false
		}
		return true, newData, false
	})
}
