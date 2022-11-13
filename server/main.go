package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
)

var (
	flagInstall = flag.Bool("install", false, "Install service")
	flagAdd     = flag.Bool("add", false, "Add user")
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
		runCommnd("enable service", "sudo", "systemctl", "enable", "ssh-auth.service")
		runCommnd("start service", "sudo", "systemctl", "start", "ssh-auth.service")
	}
	for {

	}
}
