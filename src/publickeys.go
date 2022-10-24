package main

import (
	"fmt"
	"github.com/liushuochen/gotable"
	"io"
	"strconv"
	"strings"
)

func copyPublicKeysWithRemote(r Remote, name []string, info bool) {
	// When "info" is true, the function is called concurrently, so print log only one time.
	sftpClient := connect(r, info)
	defer func() { fatalErr(sftpClient.Close()) }()
	remotePath := "/home/" + r.username + "/.ssh/authorized_keys"
	if r.username == "root" {
		remotePath = "/root/.ssh/authorized_keys"
	}
	srcFile, err := sftpClient.Open(remotePath)
	fatalErr(err)
	defer func() { fatalErr(srcFile.Close()) }()
	// read remote file and check
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
	// insert public key when it is not exists
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
	// write new content to remote file
	_, err = srcFile2.Write([]byte(newContent))
	fatalErr(err)
	if info {
		if cnt <= 1 {
			fmt.Printf("Successfully copied %d key.\n", cnt)
		} else {
			fmt.Printf("Successfully copied %d keys.\n", cnt)
		}
		succeed := 0
		for _, e := range name {
			if insertLink(r.id, e) {
				succeed += 1
			}
		}
		if succeed <= 1 {
			fmt.Printf("Successfully authorized user %s to log in server %v.\n", name[0], r)
		} else {
			fmt.Printf("Successfully authorized %d users to log in server %v.\n", succeed, r)

		}
	} else {
		if cnt <= 1 {
			fmt.Printf("Successfully copied %d key to %v.\n", cnt, r)
		} else {
			fmt.Printf("Successfully copied %d keys to %v.\n", cnt, r)
		}
	}
}

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
	copyPublicKeysWithRemote(r, name, true)
}

func syncPublicKeys(serversName []string) {
	servers := make([]Remote, 0)
	if len(serversName) == 0 {
		servers = getServers()
	} else {
		for _, e := range serversName {
			r, err := findServer(e)
			fatalErr(err)
			servers = append(servers, r)
		}
	}
	cnt := len(servers)
	cnt2 := cnt
	// maximum concurrency: 10
	next := make(chan int, 10)
	stop := make(chan int)
	for _, r := range servers {
		// wait to execution
		next <- 1
		go func(des Remote) {
			// sqlite support query concurrently, so query links in coroutines
			copyPublicKeysWithRemote(des, findLinks(des.id), false)
			<-next
			// continue main coroutines when all jobs are finished
			cnt -= 1
			if cnt <= 0 {
				stop <- 1
			}
		}(r)
	}
	// wait for jobs are finished
	<-stop
	if cnt2 <= 1 {
		fmt.Printf("Successfully synchronized %d server.\n", cnt2)
	} else {
		fmt.Printf("Successfully synchronized %d servers.\n", cnt2)
	}
}

func showLinks() {
	table, _ := gotable.Create("ID", "User", "Server")
	links := getLinks()
	for k, v := range links {
		for _, e := range v {
			table.AddRow([]string{strconv.Itoa(e.id), k, e.String()})
		}
	}

	fmt.Print(table)
}
