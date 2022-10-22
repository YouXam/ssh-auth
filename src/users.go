package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func addUser(name string, path []string) {
	result := insertUser(name)
	failed := 0
	duplicated := 0
	for _, e := range path {
		file, err := os.Open(e)
		fatalErr(err)
		content, err := io.ReadAll(file)
		fatalErr(err)
		key := strings.Trim(string(content), "\n")
		if !strings.HasPrefix(key, "ssh") {
			fmt.Printf("%s does not appear to be a public key file.\n", e)
			failed += 1
			continue
		}
		if !insertPublicKeys(name, key) {
			duplicated += 1
		}
		fatalErr(file.Close())
	}
	var length = len(path)
	if result {
		fmt.Printf("Successfully added %s.\n", name)
	}
	if length-failed-duplicated == 1 {
		fmt.Printf("Successfully added 1 key.\n")
	} else if length-failed-duplicated > 1 {
		fmt.Printf("Successfully added %d keys.\n", length)
	}
	if duplicated == 1 {
		fmt.Printf("%d key has already been added.\n", failed)
	} else if duplicated > 1 {
		fmt.Printf("%d keys have already been added.\n", failed)
	}
}
