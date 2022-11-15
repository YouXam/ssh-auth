package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
)

var (
	flagInstall   = flag.Bool("install", false, "Install service")
	flagUninstall = flag.Bool("uninstall", false, "Uninstall service")
	flagAdd       = flag.Bool("add", false, "Add user")
	flagClientKey = flag.String("client-key", "", "Client public key")
	flagHash      = flag.String("hash", "", "Client public key hash")
)

func runCommnd(info string, command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println(info)
	err := cmd.Run()
	if err != nil {
		log.Fatalln("Failed to run command:", err)
	}
}

func main() {
	flag.Parse()
	if *flagInstall {
		runCommnd("install binary to /usr/local/bin", "sudo", "cp", os.Args[0], "/usr/local/bin/ssh-auth-server")
		err := os.WriteFile("ssh-auth-server.service", []byte(serviceFile), 0644)
		if err != nil {
			log.Fatalln("Failed to write service file:", err)
		}
		runCommnd("install service file to /etc/systemd/system", "sudo", "cp", "ssh-auth-server.service", "/usr/lib/systemd/system/")
		err = os.Remove("ssh-auth-server.service")
		if err != nil {
			log.Println("Failed to delete service file:", err)
		}
		runCommnd("reload daemon", "sudo", "systemctl", "daemon-reload")
		runCommnd("enable service", "sudo", "systemctl", "enable", "ssh-auth-server.service")
		runCommnd("start service", "sudo", "systemctl", "start", "ssh-auth-server.service")
	} else if *flagUninstall {
		runCommnd("stop service", "sudo", "systemctl", "stop", "ssh-auth-server.service")
		runCommnd("disable service", "sudo", "systemctl", "disable", "ssh-auth-server.service")
		runCommnd("remove service file", "sudo", "rm", "/usr/lib/systemd/system/ssh-auth-server.service")
		runCommnd("remove binary", "sudo", "rm", "/usr/local/bin/ssh-auth-server")
	} else {
		initDatabase()
		if *flagAdd {
			if *flagClientKey == "" {
				log.Fatalln("Missing client public key")
			}
			if *flagHash == "" {
				log.Fatalln("Missing client UUID")
			}
			clientKey, err := os.ReadFile(*flagClientKey)
			if err != nil {
				log.Fatalln("Failed to read client public key:", err)
			}
			err = insertClientPublicKey(*flagHash, string(clientKey))
			fatalErr(err)
		} else {
			go watch()
			server()
			// sign := make(chan os.Signal)
			// signal.Notify(sign)
			// for {
			// 	if <-sign == os.Interrupt {
			// 		fmt.Println("Interrupted")
			// 		break
			// 	}
			// }
		}
	}
}
