package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	flagInstall   = flag.Bool("install", false, "Install service")
	flagUninstall = flag.Bool("uninstall", false, "Uninstall service")
	flagAdd       = flag.Bool("add", false, "Add user")
	flagClientKey = flag.String("client-key", "", "Client public key")
	flagHash      = flag.String("hash", "", "Client public key hash")
	flagUser      = flag.String("user", "", "Username")
)

func runCommand(info string, command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println(info)
	err := cmd.Run()
	if err != nil {
		log.Fatalln("Failed to run command:", err)
	}
}

func httpAdd(clientKey []byte, username string) error {
	data := ClientData{
		Hash:      *flagHash,
		PublicKey: string(clientKey),
		Username:  string(username),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://localhost:22222/client", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(strings.TrimSpace(string(respText)))
	return nil
}

func main() {
	flag.Parse()
	if *flagInstall {
		runCommand("install binary to /usr/local/bin", "sudo", "cp", os.Args[0], "/usr/local/bin/ssh-auth-server")
		err := os.WriteFile("ssh-auth-server.service", []byte(serviceFile), 0644)
		if err != nil {
			log.Fatalln("Failed to write service file:", err)
		}
		runCommand("install service file to /etc/systemd/system", "sudo", "cp", "ssh-auth-server.service", "/usr/lib/systemd/system/")
		err = os.Remove("ssh-auth-server.service")
		if err != nil {
			log.Println("Failed to delete service file:", err)
		}
		runCommand("reload daemon", "sudo", "systemctl", "daemon-reload")
		runCommand("enable service", "sudo", "systemctl", "enable", "ssh-auth-server.service")
		runCommand("start service", "sudo", "systemctl", "start", "ssh-auth-server.service")
	} else if *flagUninstall {
		runCommand("stop service", "sudo", "systemctl", "stop", "ssh-auth-server.service")
		runCommand("disable service", "sudo", "systemctl", "disable", "ssh-auth-server.service")
		runCommand("remove service file", "sudo", "rm", "/usr/lib/systemd/system/ssh-auth-server.service")
		runCommand("remove binary", "sudo", "rm", "/usr/local/bin/ssh-auth-server")
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
			err = httpAdd(clientKey, *flagUser)
			if err != nil {
				log.Println("Server not running, adding to database directly")
				err = insertClientPublicKey(*flagHash, string(clientKey), *flagUser)
				if err != nil {
					log.Fatalln("Failed to insert client public key:", err)
				}
			}
			log.Println("Client public key added")
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
