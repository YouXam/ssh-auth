package main

import (
	"fmt"
	"io"
	"strings"
)

func copyPublicKeys(destination string, port int, name []string, usePassword bool, privateKey string) {
	r, err := findServer(destination)
	if err != nil {
		r = parseRemote(port, destination, usePassword, privateKey)
	} else if r.password == "" && usePassword {
		r = parseRemote(r.port, destination, true, "")
	} else if privateKey != "" {
		r = parseRemote(r.port, destination, false, privateKey)
	}
	if port != 0 && port != r.port {
		r.port = port
	}
	sftpClient := connect(r)
	defer func() { fatalErr(sftpClient.Close()) }()
	fatalErr(err)
	remotePath := "/home/" + r.username + "/.ssh/authorized_keys"
	if r.username == "root" {
		remotePath = "/root/.ssh/authorized_keys"
	}
	srcFile, err := sftpClient.Open(remotePath)
	fatalErr(err)
	defer func() { fatalErr(srcFile.Close()) }()

	data, err := io.ReadAll(srcFile)
	fatalErr(err)
	content := string(data)
	publicKeys := make([]string, 0)
	publicKeysTmp := strings.Split(content, "\n")
	publicKeysMap := map[string]bool{}
	for _, e := range publicKeysTmp {
		key := strings.TrimSpace(e)
		if key != "" {
			publicKeys = append(publicKeys, key)
			publicKeysMap[key] = true
		}
	}
	cnt := 0
	for _, e := range name {
		res := findPublicKeys(e)
		for _, h := range res {
			if _, ok := publicKeysMap[h]; !ok {
				publicKeys = append(publicKeys, h)
				cnt += 1
			}
		}
	}
	newContent := strings.Join(publicKeys, "\n")
	srcFile2, err := sftpClient.Create(remotePath)
	fatalErr(err)
	defer func() { fatalErr(srcFile2.Close()) }()

	_, err = srcFile2.Write([]byte(newContent))
	fatalErr(err)

	fatalErr(err)
	if cnt <= 1 {
		fmt.Printf("Successfully copied %d key.\n", cnt)
	} else {
		fmt.Printf("Successfully copied %d keys.\n", cnt)
	}
}

func syncPublicKeys() {

}
