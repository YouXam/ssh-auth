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
	for _, e := range path {
		file, err := os.Open(e)
		checkErr(err)
		content, err := io.ReadAll(file)
		checkErr(err)
		key := strings.Trim(string(content), "\n")
		if !insertPublicKeys(name, key) {
			failed += 1
		}
		checkErr(file.Close())
	}
	var length = len(path)
	if result {
		fmt.Printf("Successfully added %s.\n", name)
	}
	if length-failed == 1 {
		fmt.Printf("Successfully added 1 key.\n")
	} else if length-failed > 1 {
		fmt.Printf("Successfully added %d keys.\n", length)
	}
	if failed == 1 {
		fmt.Printf("%d key has already been added.\n", failed)
	} else if failed > 1 {
		fmt.Printf("%d keys have already been added.\n", failed)
	}
}
