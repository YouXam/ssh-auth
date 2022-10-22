package main

import (
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type Remote struct {
	username   string
	hostname   string
	port       int
	password   string
	privateKey string
	servername string
}

func (t Remote) toString() string {
	if t.servername != "" {
		return t.servername
	}
	return t.username + "@" + t.hostname + ":" + strconv.Itoa(t.port)
}

func parseRemote(port int, destination string, usePassword bool, publicKeyPath string) Remote {
	if port <= 0 {
		port = 22
	}
	pos := strings.Index(destination, "@")
	cur, err := user.Current()
	fatalErr(err)
	username := cur.Name
	hostname := destination
	password := ""
	if pos > 0 {
		res := strings.SplitN(destination, "@", 2)
		username = res[0]
		hostname = res[1]
	}
	if usePassword || publicKeyPath == "" {
		usePassword = true
		fmt.Printf("%s@%s:%d Password:", username, hostname, port)
		passwordByte, err := gopass.GetPasswdMasked()
		password = string(passwordByte)
		fatalErr(err)
	}
	key := ""
	if publicKeyPath != "" {
		file, err := os.Open(publicKeyPath)
		fatalErr(err)
		content, err := io.ReadAll(file)
		fatalErr(err)
		key = strings.Trim(string(content), "\n")
		if !strings.HasPrefix(key, "-----BEGIN OPENSSH PRIVATE KEY-----") {
			fmt.Printf("%s does not appear to be a private key file.\nExited.\n", publicKeyPath)
			os.Exit(1)
		}
		fatalErr(file.Close())
	}
	return Remote{username, hostname, port, password, key, *flagServerName}
}

func connect(r Remote) *sftp.Client {
	auth := make([]ssh.AuthMethod, 0)
	if r.privateKey != "" {
		key, err := ssh.ParsePrivateKey([]byte(r.privateKey))
		switch err.(type) {
		case *ssh.PassphraseMissingError:
			fmt.Printf("Private Key Passphrase:")
			password, err := gopass.GetPasswdMasked()
			fatalErr(err)
			key, err = ssh.ParsePrivateKeyWithPassphrase([]byte(r.privateKey), password)
			fatalErr(err)
		default:
			fatalErr(err)
		}
		auth = append(auth, ssh.PublicKeys(key))
	}
	if r.password != "" {
		auth = append(auth, ssh.Password(r.password))
	}
	clientConfig := &ssh.ClientConfig{
		User:            r.username,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", r.hostname, r.port)

	fmt.Printf("Connecting %s@%s:%d...", r.username, r.hostname, r.port)
	sshClient, err := ssh.Dial("tcp", addr, clientConfig)
	fatalErr(err)
	sftpClient, err := sftp.NewClient(sshClient)
	fatalErr(err)
	fmt.Println("ok")
	return sftpClient
}
