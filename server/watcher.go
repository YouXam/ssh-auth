package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher

func getUsernameFromPath(path string) string {
	if strings.HasPrefix("/root/", path) {
		return "root"
	}
	// /home/username/.ssh/authorized_keys
	return strings.Split(path, "/")[2]
}

func updateAuthorizedKeys(username string) {
	modifyTag := false
	// Read the file
	path := getAuthorizedKeysPath(username)
	data, err := os.ReadFile(path)
	if err != nil {
		// if the file does not exist, create it
		if os.IsNotExist(err) {
			log.Println("Create:", path)
			data = []byte("")
		} else {
			log.Println(err)
			return
		}
	}
	publicKeys, err := getSSHPublicKeysByUsername(username)
	publicKeysMap := make(map[string]bool)
	if err != nil {
		log.Println(err)
		return
	}
	// Split lines and check if the publicKey is already in the file
	linesB := strings.Split(string(data), "\n")
	lines := make([]string, 0)
	for _, line := range linesB {
		if line != "" {
			lines = append(lines, line)
			publicKeysMap[line] = true
		}
	}
	for _, publicKey := range publicKeys {
		if _, ok := publicKeysMap[publicKey]; !ok {
			modifyTag = true
			lines = append(lines, strings.TrimSpace(publicKey))
		}
	}
	if modifyTag {
		// Write the file
		modifyTag = true
		err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
		if err != nil {
			log.Println(err)
		}
		log.Println("Update:", path)
	}
}

var fileMap = make(map[string]bool)

func watch() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
	}
	defer watcher.Close()
	usernames := getUsernames()
	for _, username := range usernames {
		updateAuthorizedKeys(username)
		path := getAuthorizedKeysPath(username)
		err = watcher.Add(path)
		if err != nil {
			log.Println(err)
		} else {
			fileMap[path] = true
			log.Println("Watching", path)
		}
	}
	log.Println("Start watching authorized_keys files")
	timeTags := make(map[string]int64)
	filesOp := make(map[string]fsnotify.Op)
	for {
		select {
		case ev := <-watcher.Events:
			{
				if _, ok := fileMap[ev.Name]; ok {
					// Delay 200ms to wait for the file to be written
					var timeTag int64
					if t, ok := timeTags[ev.Name]; ok {
						timeTag = t
					}
					filesOp[ev.Name] = ev.Op
					if time.Duration(time.Now().UnixMilli()-timeTag) > 200*time.Millisecond {
						timeTags[ev.Name] = time.Now().UnixMilli()
						go func() {
							time.Sleep(200 * time.Millisecond)
							log.Printf("File modified[%v]: %s", filesOp[ev.Name], ev.Name)
							username := getUsernameFromPath(ev.Name)
							updateAuthorizedKeys(username)
							watcher.Add(ev.Name)
							timeTags[ev.Name] = 0
						}()
					}
				}
			}
		case err := <-watcher.Errors:
			log.Println(err)
		}
	}
}
