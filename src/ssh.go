package main

import (
	"bytes"
	"crypto/sha256"
	"embed"
	_ "embed"
	"encoding/hex"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type Remote struct {
	username    string
	hostname    string
	port        int
	password    string
	privateKey  string
	servername  string
	id          int
	isInstalled int
}

//go:embed bin
var sshAuthServerBin embed.FS

var archMap = map[string]string{
	"x86_64":    "amd64",
	"aarch64":   "arm64",
	"armv7l":    "arm",
	"armv6l":    "arm",
	"armv5tel":  "arm",
	"armv5tejl": "arm",
	"armv4tl":   "arm",
	"armv4t":    "arm",
	"i686":      "386",
	"i386":      "386",
	"i586":      "386",
	"i486":      "386",
}

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

func checkDaemon(hostname string) int {
	resp, err := http.Get("http://" + hostname + ":22222/ping")
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	respText, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}
	if strings.TrimSpace(string(respText)) == "ssh-auth-server.pong" {
		fmt.Println("Successfully installed daemon.")
		return 1
	}
	fmt.Println("Failed to installed daemon.")
	return 0
}

func addClient(s *ssh.Client, r Remote) int {
	sftpClient, err := sftp.NewClient(s)
	fatalErrRemote(r, err)
	// check client key
	_, clientPublicKey := checkKeyPair(r.username + "@" + r.hostname)
	// Upload clientPublicKey to remote: /tmp/client_public_key.<random string>
	path := "/tmp/client_public_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".key"
	writeRemote(sftpClient, path, []byte(clientPublicKey))
	fmt.Println("Uploaded clientPublicKey to remote: " + path)
	// Add Client
	hash := sha256.Sum256([]byte(clientPublicKey))
	hashHexText := hex.EncodeToString(hash[:])
	result, err := runSSHCommand(s, "ssh-auth-server -add -user "+r.username+" -client-key "+path+" -hash "+hashHexText, r)
	fmt.Println(result)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Failed to add client.")
		return 0
	}
	return 1
}

func installServer(s *ssh.Client, r Remote) int {
	if checkDaemon(r.hostname) == 1 {
		if r.isInstalled == 1 {
			return 1
		}
		return addClient(s, r)
	}
	// let users choose if install server
	fmt.Print("\nDo you want to install daemon on remote host? [Y/n]")
	var input string
	_, _ = fmt.Scanln(&input)
	if input == "n" || input == "N" {
		return 0
	}

	// test permission (root or sudo)
	password := "test"
	sudoResult, _ := runSSHCommand(s, "echo -n \""+password+"\" | sudo -S -p \"\" whoami", r)
	if strings.TrimSpace(sudoResult) != "root" {
		// permission denied
		for {
			fmt.Printf("[sudo] password for " + r.username + ": ")
			passwordByte, err := gopass.GetPasswdMasked()
			fatalErr(err)
			password = strings.TrimSpace(string(passwordByte))
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

	arch, err := runSSHCommand(s, "uname -m", r)
	fmt.Println("Remote arch: " + strings.TrimSpace(arch))
	data, err := sshAuthServerBin.ReadFile("bin/ssh-auth-server-linux-" + archMap[strings.TrimSpace(arch)])
	if err != nil {
		fmt.Println("Unsupported arch: " + arch)
		os.Exit(1)
	}

	fmt.Print("Uploading executable to remote: /tmp/ssh-auth-server/ssh-auth-server...")
	// Upload executable to remote: /tmp/ssh-auth-server/ssh-auth-server
	writeRemote(sftpClient, "/tmp/ssh-auth-server/ssh-auth-server", data)
	fmt.Println("ok")

	// Change permission
	_, _ = runSSHCommand(s, "chmod +x /tmp/ssh-auth-server/ssh-auth-server", r)

	// Install daemon
	_, _ = runSSHCommand(s, "echo -n \""+password+"\" | sudo -S -p \"\" /tmp/ssh-auth-server/ssh-auth-server --install", r)
	fmt.Println("Waiting for daemon to start...")
	time.Sleep(1 * time.Second)
	ok := checkDaemon(r.hostname)
	if ok == 0 {
		return 0
	}
	return addClient(s, r)
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
	if info {
		fmt.Println("ok")
	}
	return
}

func connect(r Remote, info bool) *sftp.Client {
	sshClient := connectSSH(r, info)
	sftpClient, err := sftp.NewClient(sshClient)
	fatalErrRemote(r, err)
	return sftpClient
}
