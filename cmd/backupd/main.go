package main

import (
	"easy_backup/internal/archiver"
	"easy_backup/utils"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/matryer/filedb"
	"github.com/joho/godotenv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type path = utils.Path
var getAbsPath = utils.GetAbsPath
var zipper archiver.Archiver

func main() {
	var fatalErr error
	defer func() {
		if fatalErr != nil {
			log.Println(fatalErr)
		}
	}()

	workingDir, _ := os.Getwd()
	defaultDbPath := filepath.Join(workingDir, "data")

	var (
		interval = *flag.Duration("interval", 10*time.Second, "Backup cycle: Interval between archive process")
		backupTo = *flag.String("archive", "backups", "Path to archive location")
		dbPath   = *flag.String("db", defaultDbPath, "Filesystem DB storing paths of files to backup")
		once = flag.Bool("once", false, "Use to run backup once")
		useCloudStorage = flag.Bool("cloud", true, "stores files to AWS S3 if true.\nTo store in local directory, change to false")
	)
	flag.Parse()

	db, err := filedb.Dial(getAbsPath(dbPath))
	if err != nil {
		fatalErr = err
		return
	}
	defer db.Close()

	if (*useCloudStorage){
		LoadEnv()

		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(GetEnv("AWSRegion")),
			Credentials: credentials.NewStaticCredentials(GetEnv("AWSAccessKeyId"), GetEnv("AWSSecretAccessKey"), ""),
		})
		if err != nil {
			fatalErr = err
			return
		}

		zipper = archiver.NewS3Archiver(GetEnv("S3Bucket"), sess)
	} else {
		zipper = archiver.LocalZipper
	}

	backupHandler := &Handler{
		Paths:       make(map[string]string),
		Archiver:    zipper,
		Destination: backupTo,
	}

	pathCollection, err := db.C(utils.PathFileName)
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
		fatalErr = errors.New("no path: add atleast one path using the backup management tool")
		return
	}

	runBackupCycle(backupHandler, pathCollection)

	if (*once) {
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-time.After(interval):
			runBackupCycle(backupHandler, pathCollection)
		case <-signalChan:
			log.Println("\nstopping...")
			return
		}
	}
}

func runBackupCycle(handler *Handler, collection *filedb.C) {
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
			log.Println("Failed to parse data to json (skipping...)", err)
			return true, data, false
		}

		path.Hash = handler.Paths[path.Path]
		newData, err := json.Marshal(path)

		if err != nil {
			log.Println("Failed to parse path struct to json (skipping...)", err)
			return true, data, false
		}

		return true, newData, false
	})
}


func GetEnv(key string) string {
	return os.Getenv(key)
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}