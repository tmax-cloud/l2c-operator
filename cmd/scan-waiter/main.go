package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"github.com/tmax-cloud/l2c-operator/pkg/apis"
)

func main() {
	arg := os.Args[1:]
	if len(arg) != 1 {
		log.Println("should have exact one argument [wait|ok|fail]")
		os.Exit(1)
	}

	switch arg[0] {
	case "wait":
		watchDir()
	case string(apis.ScanResultOk):
		createFile(apis.ScanResultOk)
	case string(apis.ScanResultFail):
		createFile(apis.ScanResultFail)
	default:
		log.Println("first arg should be one of [wait|ok|fail]")
		os.Exit(1)
	}
}

func watchDir() {
	if !utils.DirExists(apis.DirName) {
		log.Printf("Creating directory %s\n", apis.DirName)
		if err := os.Mkdir(apis.DirName, 0644); err != nil {
			log.Fatalf("cannot create dir %s", apis.DirName)
		}
	}

	if utils.FileExists(apis.DirName + "/" + string(apis.ScanResultOk)) {
		log.Printf("File %s found", apis.DirName+"/"+string(apis.ScanResultOk))
		os.Exit(0)
	} else if utils.FileExists(apis.DirName + "/" + string(apis.ScanResultFail)) {
		log.Printf("File %s found", apis.DirName+"/"+string(apis.ScanResultFail))
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		log.Printf("Watching directory %s\n", apis.DirName)
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("created file :", event.Name)
					if strings.HasSuffix(event.Name, string(apis.ScanResultOk)) {
						os.Exit(0)
					} else if strings.HasSuffix(event.Name, string(apis.ScanResultFail)) {
						os.Exit(1)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	if err = watcher.Add(apis.DirName); err != nil {
		log.Fatal(err)
	}

	<-done
}

func createFile(result apis.ScanResult) {
	if !utils.DirExists(apis.DirName) {
		log.Printf("Creating directory %s\n", apis.DirName)
		if err := os.Mkdir(apis.DirName, 0644); err != nil {
			log.Fatalf("cannot create dir %s", apis.DirName)
		}
	}

	filePath := apis.DirName + "/" + string(result)
	_, err := fmt.Fprintf(os.Stdout, "Writing file %s\n", filePath)
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filePath, []byte("\n"), 0755); err != nil {
		log.Fatal(err)
	}
}
