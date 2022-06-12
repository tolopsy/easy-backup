package main

import (
	backup "easy_backup"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/matryer/filedb"
)

// path is same as in backup program.
// It is duplicated thus rather than placed in a different package
// to avoid extra dependency injection
// with a little overhead trade-off (which in this case is the duplication)
type path struct {
	Path string
	Hash string
}

func (p path) String() string {
	return fmt.Sprintf("%s [%s]", p.Path, p.Hash)
}

func main() {
	var fatalErr error
	defer func() {
		if fatalErr != nil {
			log.Println(fatalErr)
		}
	}()

	var (
		interval = flag.Duration("interval", 10*time.Second, "Backup cycle: Interval between archive process")
		backupTo = flag.String("backup_dir", "backups", "Path to archive location")
		dbPath   = flag.String("db", "../backup/data", "Filesystem DB storing paths of files to backup")
	)
	flag.Parse()

	backupHandler := &backup.Handler{
		Paths:       make(map[string]string),
		Archiver:    backup.ZIP,
		Destination: *backupTo,
	}

	db, err := filedb.Dial(*dbPath)
	if err != nil {
		fatalErr = err
		return
	}
	defer db.Close()

	pathCollection, err := db.C("paths")
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
	counter, err := handler.Run()
	if err != nil {
		log.Fatalln("Failed to backup - ", err)
	}

	if counter == 0 {
		log.Println("\tNo changes!")
		return
	}

	log.Printf("\t%d Directories backed up", counter)
	var path path
	collection.SelectEach(func(_ int, data []byte) (bool, []byte, bool) {
		if err = json.Unmarshal(data, &path); err != nil {
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
