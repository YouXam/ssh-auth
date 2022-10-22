package main

import (
	"fmt"
)

func addServer(destination string, port int, usePassword bool, publicKeyPath string, serverName string) {
	r := parseRemote(port, destination, usePassword, publicKeyPath)
	fmt.Println("Testing ssh connection...")
	connect(r)
	if insertServer(r.hostname, r.port, r.username, serverName, r.password, r.privateKey) {
		fmt.Printf("Successfully added server %s.\n", r.toString())
	} else {
		fmt.Printf("Successfully modified server %s.\n", r.toString())
	}
}
