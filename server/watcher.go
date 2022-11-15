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
		log.Println(err)
		return
	}
	publicKeys, err := getSSHPublicKeysByUsername(username)
	publicKeysMap := make(map[string]bool)
	if err != nil {
		log.Println(err)
		return
	}
	// Splitlines and check if the publicKey is already in the file
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
	}
}

func watch() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
	}
	defer watcher.Close()
	usernames := getUsernames()
	fileMap := make(map[string]bool)
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
	go func() {
		// every 10 minutes, check if there are new username.
		time.Sleep(10 * time.Minute)
		usernames := getUsernames()
		for _, username := range usernames {
			path := getAuthorizedKeysPath(username)
			if _, ok := fileMap[path]; !ok {
				err = watcher.Add(path)
				if err != nil {
					log.Println(err)
				} else {
					fileMap[path] = true
				}
			}
		}
	}()
	log.Println("Start watching authorized_keys files")
	for {
		select {
		case ev := <-watcher.Events:
			{
				log.Printf("%s %v, try to update it\n", ev.Name, ev.Op)
				username := getUsernameFromPath(ev.Name)
				updateAuthorizedKeys(username)
				watcheList := watcher.WatchList()
				flag := false
				for _, path := range watcheList {
					if path == ev.Name {
						flag = true
					}
				}
				if !flag {
					watcher.Add(ev.Name)
				}
			}
		case err := <-watcher.Errors:
			log.Println(err)
		}
	}
}
