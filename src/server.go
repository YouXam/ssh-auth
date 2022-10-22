package main

import (
	"fmt"
	"github.com/howeyc/gopass"
	"io"
	"os"
	"os/user"
	"strings"
)

func addServer(destination string, port int, usePassword bool, publicKeyPath string, serverName string) {
	pos := strings.Index(destination, "@")
	cur, err := user.Current()
	checkErr(err)
	username := cur.Name
	hostname := destination
	password := ""
	if pos > 0 {
		res := strings.SplitN(destination, "@", 2)
		username = res[0]
		hostname = res[1]
	}
	if usePassword {
		fmt.Printf("Password:")
		passwordByte, err := gopass.GetPasswdMasked()
		password = string(passwordByte)
		checkErr(err)
	}
	key := ""
	if publicKeyPath != "" {
		file, err := os.Open(publicKeyPath)
		checkErr(err)
		content, err := io.ReadAll(file)
		checkErr(err)
		key = strings.Trim(string(content), "\n")
		checkErr(file.Close())
	}
	serverId := destination
	if serverName != "" {
		serverId = serverName
	}
	if insertServer(hostname, port, username, serverName, password, key) {
		fmt.Printf("Successfully added server %s.\n", serverId)
	} else {
		fmt.Printf("Successfully modified server %s.\n", serverId)
	}
	// TODO: test ssh connection
}
