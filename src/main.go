package main

import (
	"flag"
	"fmt"
	"os"
)

const VERSION = "0.0.1"

var (
	flagHelp       = flag.Bool("help", false, "print more detailed help information")
	flagVersion    = flag.Bool("version", false, "print ssh-auth version")
	flagPassword   = flag.Bool("P", false, "use password to connect server")
	flagPrivateKey = flag.String("i", "", "private key to connect server")
	flagServerName = flag.String("n", "", "server name")
	flagPort       = flag.Int("p", 0, "server port")
)

func Usage() {
	fmt.Println(`ssh-auth - manage multi-user logins on remote machines using imported keys
SYNOPSIS
	ssh-auth --version
	ssh-auth --help
	ssh-auth <command> [<args>]
COMMANDS
	ssh-auth user <name> [path [path2 [path3 ...]]]
		Import public keys, and add user when user is not exists.
	ssh-auth server [-p port] [-P] [-i path] [-n name] [user@]hostname
		Add server.
			-p: server port, default value is 22.
			-P: use password to connect server, it will be saved in clear text.
			-i: private key to connect server, it will be saved in clear text.
			-n: name of server.
	ssh-auth copy [-p port] [-P] [-i path] <user> [user2 [user3 ...]] <servername|[username@]hostname> 
		Copy public keys to remote machine.
			-p: server port, default value is 22.
			-P: use password to connect server, it will not be saved.
			-i: private key to connect server, it will not be saved.
	ssh-auth sync [servername|[username@]hostname] [servername|[username@]hostname] ...
		Synchronize the public key of specified servers or all servers.`)
}

func main() {
	initDatabase()

	flag.Parse()
	if *flagVersion {
		fmt.Println(VERSION)
		return
	}
	if *flagHelp {
		Usage()
		return
	}
	if len(os.Args) < 2 {
		fmt.Println("No subcommand found.")
		Usage()
		os.Exit(1)
	}
	command := os.Args[1]
	os.Args = os.Args[1:]
	// parse args again to skip subcommand
	flag.Parse()
	args := flag.Args()
	switch command {
	case "user":
		if len(args) < 1 {
			fmt.Println("Missing necessary argument: name.")
			Usage()
			os.Exit(1)
		}
		addUser(args[0], args[1:])
	case "server":
		if len(args) < 1 {
			fmt.Println("Missing necessary argument: destination.")
			Usage()
			os.Exit(1)
		}
		addServer(args[0], *flagPort, *flagPassword, *flagPrivateKey, *flagServerName)
	case "copy":
		if len(args) < 2 {
			fmt.Println("Missing necessary argument.")
			Usage()
			os.Exit(1)
		}
		copyPublicKeys(args[len(args)-1], *flagPort, args[:len(args)-1], *flagPassword, *flagPrivateKey)
	case "sync":
		syncPublicKeys(args)
	default:
		fmt.Println("Invalid subcommand.")
	}
}
