package main

import (
	"fmt"
	"github.com/liushuochen/gotable"
	"strconv"
)

func addServer(destination string, port int, usePassword bool, publicKeyPath string, serverName string) {
	r := parseRemote(port, destination, usePassword, publicKeyPath)
	fmt.Println("Testing ssh connection...")
	sshClient := connectSSH(r, true)
	isInstalled := installServer(sshClient, r)
	if insertServer(r.hostname, r.port, r.username, serverName, r.password, r.privateKey, isInstalled) {
		fmt.Printf("Successfully added server %v.\n", r)
	} else {
		fmt.Printf("Successfully modified server %v.\n", r)
	}
}

func editServer(destination string, port int, usePassword bool, publicKeyPath string, serverName string) {
	r, err := findServer(destination)
	fatalErr(err)
	if port > 0 && r.port != port {
		r.port = port
	}
	if usePassword && r.password == "" || publicKeyPath != "" {
		r = parseRemote(r.port, r.username+"@"+r.hostname, usePassword, publicKeyPath)
	}
	if serverName != "" && r.servername != serverName {
		r.servername = serverName
	}
	fmt.Println("Testing ssh connection...")
	sshClient := connectSSH(r, true)
	isInstalled := installServer(sshClient, r)
	insertServer(r.hostname, r.port, r.username, serverName, r.password, r.privateKey, isInstalled)
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

func delServers(servers []string) {
	for _, e := range servers {
		r, err := findServer(e)
		if err != nil {
			fmt.Println(err)
			continue
		}
		delLinks(findLinks(r.id))
		deleteLinkByServerID(r.id)
		deleteServerByID(r.id)
		fmt.Printf("Successfully removed server %s.\n", r.String())
	}
}
