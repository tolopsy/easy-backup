package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/matryer/filedb"
)

type path struct {
	Path string
	Hash string
}

func (p path) String() string {
	return fmt.Sprintf("%s [%s]", p.Path, p.Hash)
}

func main() {
	var fatalErr error
	pathFileName := "Paths"
	dbDefaultPath := "./data"
	defer func() {
		if fatalErr != nil {
			flag.PrintDefaults()
			log.Fatalln(fatalErr)
		}
	}()

	dbPath := flag.String("db", dbDefaultPath, "Filesystem DB storing paths of files to backup")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fatalErr = errors.New("invalid usage: arguments must be specified")
		return
	}
	db, err := filedb.Dial(*dbPath)
	if err != nil {
		fatalErr = err
		return
	}
	defer db.Close()

	pathCollection, err := db.C(pathFileName)
	if err != nil {
		fatalErr = err
		return
	}

	switch strings.ToLower(args[0]) {
	case "list":
		var path path
		err = pathCollection.ForEach(func(_ int, data []byte) bool {
			if err := json.Unmarshal(data, &path); err != nil {
				fatalErr = err
				return true
			}
			fmt.Printf("=%s\n", path)
			return false
		})
		if err != nil {
			fatalErr = err
			return
		}
	case "add":
		if len(args[1:]) == 0 {
			fatalErr = errors.New("command must specify the path to add")
			return
		}
		for _, p := range args[1:] {
			path := &path{Path: p, Hash: "Not backed up yet"}
			if err = pathCollection.InsertJSON(path); err != nil {
				fatalErr = err
				return
			}
			fmt.Printf("+ %s\n", path)
		}
	case "remove":
		var path path
		err = pathCollection.RemoveEach(func(_ int, data []byte) (bool, bool){
			if err := json.Unmarshal(data, &path); err != nil {
				fatalErr = err
				return false, true
			}

			for _, pathFromArgs := range args[1:] {
				if path.Path == pathFromArgs {
					fmt.Printf("- %s\n", path)
					return true, false
				}
			}
			return false, false
		})
		if err != nil {
			fatalErr = err
			return
		}
	}
}
