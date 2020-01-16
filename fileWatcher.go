package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func parseArguments() (fileName, podName, namespace *string, err error) {
	fileName = flag.String("file", "", "file name(required)")
	podName = flag.String("pod", "", "pod name(partial name is acceptable)(required)")
	namespace = flag.String("namespace", "", "namespace(required)")

	flag.Parse()

	if *podName == "" {
		err = errors.New("pod name not provided")
	} else if *namespace == "" {
		err = errors.New("namespace not provided")
	} else if *fileName == "" {
		err = errors.New("fileName not provided")
	}

	return
}

func deletePod(podName, namespace string) (err error) {
	cmdString := fmt.Sprintf("kubectl -n %s get pods | grep %s | awk '{print $1}' | xargs kubectl -n %s delete pod", namespace, podName, namespace)

	cmd := exec.Command("bash", "-c", cmdString)

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}

	if !strings.Contains(out.String(), "deleted") {
		err = errors.New("pod was not deleted")
	} else {
		log.Println(podName, " restarted")
	}

	return
}

func main() {

	fileName, podName, namespace, err := parseArguments()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	err = watcher.Add(*fileName)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			log.Println("event:", event)
			if event.Op&fsnotify.Remove != fsnotify.Remove {
				go func() {
					err := deletePod(*podName, *namespace)
					if err != nil {
						log.Printf("Following error occurred while restarting POD : %v\n", err)
					}
				}()
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
