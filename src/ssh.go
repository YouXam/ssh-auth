package main

import (
	"bytes"
	_ "embed"
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
	username        string
	hostname        string
	port            int
	password        string
	privateKey      string
	servername      string
	id              int
	installedDaemon int
}

////go:embed ssh-auth-server
//var sshAuthServer []byte

func (t Remote) String() string {
	if t.servername != "" {
		return t.servername + "(" + t.username + "@" + t.hostname + ":" + strconv.Itoa(t.port) + ")"
	}
	return t.username + "@" + t.hostname + ":" + strconv.Itoa(t.port)
}

func runSSHCommand(s *ssh.Client, cmd string, r Remote) (string, error) {
	var stdoutBuf bytes.Buffer
	session, err := s.NewSession()
	fatalErrRemote(r, err)
	defer func() { _ = session.Close() }()
	session.Stdout = &stdoutBuf
	err = session.Run(cmd)
	return stdoutBuf.String(), err
}

func installServer(s *ssh.Client, r Remote) {
	// let users choose if install server
	fmt.Print("\nDo you want to install daemon on remote host? [Y/n]")
	var input string
	_, _ = fmt.Scanln(&input)
	if input == "n" || input == "N" {
		return
	}
	// check client key
	_, clientPublicKey := checkKeyPair()
	// test permission (root or sudo)
	sudoResult, _ := runSSHCommand(s, "echo -n \"test\" | sudo -S -p \"\" whoami", r)
	if strings.TrimSpace(sudoResult) != "root" {
		// permission denied
		for {
			fmt.Printf("[sudo] password for " + r.username + ": ")
			passwordByte, err := gopass.GetPasswdMasked()
			fatalErr(err)
			password := strings.TrimSpace(string(passwordByte))
			sudoResult, _ = runSSHCommand(s, "echo -n \""+password+"\" | sudo -S -p \"\" whoami", r)
			if strings.TrimSpace(sudoResult) == "root" {
				break
			}
			fmt.Println(sudoResult)
			fmt.Println("Permission denied, please try again.")
		}
	}
	fmt.Println("Installing daemon...")
	sftpClient, err := sftp.NewClient(s)
	fatalErrRemote(r, err)
	_, _ = runSSHCommand(s, "rm -rf /tmp/ssh-auth-server", r)
	fatalErrRemote(r, sftpClient.Mkdir("/tmp/ssh-auth-server"))
	fmt.Println("Created directory: /tmp/ssh-auth-server")

	// Upload clientPublicKey to remote: /tmp/ssh-auth-server/client_public.key
	writeRemote(sftpClient, "/tmp/ssh-auth-server/client_public.key", []byte(clientPublicKey))
	fmt.Println("Upload clientPublicKey to remote: /tmp/ssh-auth-server/client_public.key")

	//// Upload executable to remote: /tmp/ssh-auth-server/ssh-auth-server
	//writeRemote(sftpClient, "/tmp/ssh-auth-server/ssh-auth-server", sshAuthServer)
	//fmt.Println("Upload executable to remote: /tmp/ssh-auth-server/ssh-auth-server")
	//
	//// Install daemon
	//_, err = runSSHCommand(s, "/tmp/ssh-auth-server/ssh-auth-server --install", r)
	//fatalErrRemote(r, err)
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
	r, err := findServer(username + "@" + hostname)
	if err == nil {
		return r
	}
	return Remote{username, hostname, port, password, key, *flagServerName, -1, 0}
}

func connectSSH(r Remote, info bool) (sshClient *ssh.Client) {
	auth := make([]ssh.AuthMethod, 0)
	if r.privateKey != "" {
		key, err := ssh.ParsePrivateKey([]byte(r.privateKey))
		// maybe this private key need a passphrase to decode
		switch err.(type) {
		case *ssh.PassphraseMissingError:
			fmt.Printf("Private Key Passphrase:")
			password, err := gopass.GetPasswdMasked()
			fatalErrRemote(r, err)
			key, err = ssh.ParsePrivateKeyWithPassphrase([]byte(r.privateKey), password)
			fatalErrRemote(r, err)
		default:
			fatalErrRemote(r, err)
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
	// when info is false, which means the function is not called concurrently, print more detailed information
	if info {
		fmt.Printf("Connecting %s@%s:%d...", r.username, r.hostname, r.port)
	}
	sshClient, err := ssh.Dial("tcp", addr, clientConfig)
	fatalErrRemote(r, err)
	return
}

func connect(r Remote, info bool) *sftp.Client {
	sshClient := connectSSH(r, info)
	sftpClient, err := sftp.NewClient(sshClient)
	fatalErrRemote(r, err)
	if info {
		fmt.Println("ok")
	}
	return sftpClient
}
