package main

import (
	"flag"
	"fmt"
	"os"
)

const VERSION = "0.0.3"

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
	ssh-auth user add <name> [path [path2 [path3 ...]]]
		Import public keys, and add user when user is not exists.
	ssh-auth user show
		Display all users
	ssh-auth user rm name [name2 [name3 ...]]]
		Remove user.
	ssh-auth server add [-p port] [-P] [-i path] [-n name] [user@]hostname
		Add server.
			-p: server port, default value is 22.
			-P: use password to connect server, it will be saved in clear text.
			-i: private key to connect server, it will be saved in clear text.
			-n: name of server.
	ssh-auth server edit [-p port] [-P] [-i path] [-n newName] <servername|[username@]hostname>
		Edit server.
			-p: server port, default value is 22.
			-P: use password to connect server, it will be saved in clear text.
			-i: private key to connect server, it will be saved in clear text.
			-n: name of server.
	ssh-auth server show
		Display all servers.
	ssh-auth server rm <servername|[username@]hostname> [servername|[username@]hostname] ...
		Remove server.
	ssh-auth auth add [-p port] [-P] [-i path] <servername|[username@]hostname> <user> [user2 [user3 ...]]
		Add authorization, and copy public keys to remote machine.
			-p: server port, default value is 22.
			-P: use password to connect server, it will not be saved.
			-i: private key to connect server, it will not be saved.
	ssh-auth auth show
		Display all authorization.
	ssh-auth auth rm id [id2 [id3 ...]]]
		Remove authorization.
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
	switch command {
	case "user":
		subcommand := os.Args[1]
		os.Args = os.Args[1:]
		// parse args again to skip subcommand
		flag.Parse()
		args := flag.Args()
		switch subcommand {
		case "add":
			if len(args) < 1 {
				fmt.Println("Missing necessary argument: name.")
				Usage()
				os.Exit(1)
			}
			addUser(args[0], args[1:])
		case "show":
			showUser()
		case "rm":
			if len(args) < 1 {
				fmt.Println("Missing necessary argument: username.")
				Usage()
				os.Exit(1)
			}
			delUsers(args)
		default:
			fmt.Printf("Invalid subcommand: %s.%s.\nExited.\n", command, subcommand)
			os.Exit(1)
		}
	case "server":
		subcommand := os.Args[1]
		os.Args = os.Args[1:]
		// parse args again to skip subcommand
		flag.Parse()
		args := flag.Args()
		switch subcommand {
		case "add":
			if len(args) < 1 {
				fmt.Println("Missing necessary argument: destination.")
				Usage()
				os.Exit(1)
			}
			addServer(args[0], *flagPort, *flagPassword, *flagPrivateKey, *flagServerName)
		case "rm":
			if len(args) < 1 {
				fmt.Println("Missing necessary argument: destination.")
				Usage()
				os.Exit(1)
			}
			delServers(args)
		case "show":
			showServer()
		case "edit":
			if len(args) < 1 {
				fmt.Println("Missing necessary argument: destination.")
				Usage()
				os.Exit(1)
			}
			editServer(args[0], *flagPort, *flagPassword, *flagPrivateKey, *flagServerName)
		default:
			fmt.Printf("Invalid subcommand: %s.%s.\nExited.\n", command, subcommand)
			os.Exit(1)
		}
	case "auth":
		subcommand := os.Args[1]
		os.Args = os.Args[1:]
		// parse args again to skip subcommand
		flag.Parse()
		args := flag.Args()
		switch subcommand {
		case "add":
			if len(args) < 2 {
				fmt.Println("Missing necessary argument.")
				Usage()
				os.Exit(1)
			}
			copyPublicKeys(args[0], *flagPort, args[1:], *flagPassword, *flagPrivateKey)
		case "rm":
			if len(args) < 1 {
				fmt.Println("Missing necessary argument.")
				Usage()
				os.Exit(1)
			}
			delLinksByStrings(args)
		case "show":
			showLinks()
		default:
			fmt.Printf("Invalid subcommand: %s.%s.\nExited.\n", command, subcommand)
			os.Exit(1)
		}
	case "sync":
		flag.Parse()
		syncPublicKeys(flag.Args())
	default:
		fmt.Printf("Invalid subcommand: %s.\n", command)
	}
}
