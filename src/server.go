package main

import (
	"fmt"
	"github.com/liushuochen/gotable"
	"strconv"
)

func addServer(destination string, port int, usePassword bool, publicKeyPath string, serverName string) {
	r := parseRemote(port, destination, usePassword, publicKeyPath)
	fmt.Println("Testing ssh connection...")
	connect(r, true)
	if insertServer(r.hostname, r.port, r.username, serverName, r.password, r.privateKey) {
		fmt.Printf("Successfully added server %v.\n", r)
	} else {
		fmt.Printf("Successfully modified server %v.\n", r)
	}
}

func showServer() {
	servers := getServers()
	table, _ := gotable.Create("ID", "Name", "Address", "Type")
	for _, e := range servers {
		ct := "password"
		if e.privateKey != "" {
			ct = "private key"
		}
		_ = table.AddRow([]string{strconv.Itoa(e.id), e.servername, e.username + "@" + e.hostname + ":" + strconv.Itoa(e.port), ct})
	}
	fmt.Print(table)

}
