package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/liushuochen/gotable"
	"github.com/pkg/sftp"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type OperatorData struct {
	Hash      string
	PublicKey string
	Sign      string
}

type OperatorResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// httpRequest sends a request to the server to add or delete the public key via http.
func httpRequest(hostname, sshPublicKey, privateKey, hashHexText, op string) error {
	signText, err := sign(sshPublicKey, privateKey)
	fatalErr(err)
	data := OperatorData{
		Hash:      hashHexText,
		PublicKey: sshPublicKey,
		Sign:      signText,
	}

	// convert data to json string
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://"+hostname+":22222/"+op, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// json decode
	var result OperatorResult
	err = json.Unmarshal(respText, &result)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf(result.Message)
	}
	return nil
}

// httpOperations sends multi requests to the server to add or delete the public key via httpRequest.
func httpOperations(destination string, sshPublicKeys []string, op string) (int, error) {
	if op != "add" && op != "del" {
		fatalErr(fmt.Errorf("invalid operation %s", op))
	}
	privateKey, publicKey := checkKeyPair(destination)
	hash := sha256.Sum256([]byte(publicKey))
	hashHexText := hex.EncodeToString(hash[:])
	//fmt.Println("publicKey:", publicKey)
	//fmt.Println("hash     :", hashHexText)
	success := 0
	for _, p := range sshPublicKeys {
		hostname := strings.Split(destination, "@")[1]
		err := httpRequest(hostname, p, privateKey, hashHexText, op)
		if err != nil {
			return 0, err
		} else {
			success += 1
		}
	}
	return success, nil
}

// sftpAdd adds public keys to remote server via sftp.
func sftpAdd(sshPublicKeys []string, r Remote, info bool) int {
	sftpClient := connect(r, info)
	defer func() { fatalErr(sftpClient.Close()) }()
	content := readRemote(sftpClient, r)
	publicKeysTmp := strings.Split(content, "\n")
	publicKeysMap := map[string]bool{}
	publicKeyTexts := make([]string, 0)
	for _, e := range publicKeysTmp {
		key := strings.TrimSpace(e)
		if key != "" {
			publicKeyTexts = append(publicKeyTexts, key)
			publicKeysMap[key] = true
		}
	}
	cnt := 0
	for _, e := range sshPublicKeys {
		if _, ok := publicKeysMap[e]; !ok {
			publicKeyTexts = append(publicKeyTexts, e)
			publicKeysMap[e] = true
			cnt += 1
		}
	}
	newContent := strings.Join(publicKeyTexts, "\n")
	writeRemote(sftpClient, getAuthorizedKeysPath(r.username), []byte(newContent))
	return cnt
}

// readRemote reads the content of the remote file via sftp.
func readRemote(sftpClient *sftp.Client, r Remote) string {
	remotePath := "/home/" + r.username + "/.ssh/authorized_keys"
	if r.username == "root" {
		remotePath = "/root/.ssh/authorized_keys"
	}
	srcFile, err := sftpClient.Open(remotePath)
	// if not exist, create it
	if err != nil {
		basePath := "/home/" + r.username + "/.ssh"
		if r.username == "root" {
			basePath = "/root/.ssh"
		}
		err := sftpClient.Mkdir(basePath)
		fatalErr(err)
		srcFile, err = sftpClient.Create(remotePath)
		fatalErr(err)
	}
	defer func() { fatalErr(srcFile.Close()) }()
	// readRemote remote file and check
	data, err := io.ReadAll(srcFile)
	fatalErr(err)
	content := string(data)
	return content
}

// getAuthorizedKeysPath return the path of authorized_keys file by username.
func getAuthorizedKeysPath(username string) string {
	if username == "root" {
		return "/root/.ssh/authorized_keys"
	}
	return "/home/" + username + "/.ssh/authorized_keys"
}

// writeRemote writes remote file to the remote file via sftp.
func writeRemote(sftpClient *sftp.Client, path string, content []byte) {
	srcFile2, err := sftpClient.Create(path)
	fatalErr(err)
	defer func() { fatalErr(srcFile2.Close()) }()
	// write new content to remote file
	_, err = srcFile2.Write(content)
	fatalErr(err)
}

// copyPublicKeysWithRemote copies public keys to remote server via http or sftp (http first)
func copyPublicKeysWithRemote(r Remote, name []string, info bool) {
	publicKeys := make([]string, 0)
	for _, e := range name {
		uid := findUserId(e)
		if uid < 0 {
			fmt.Printf("Can not find user %s\n", e)
		}
		res := findPublicKeys(e)
		for _, h := range res {
			publicKeys = append(publicKeys, h)
		}
	}
	cnt := 0
	// When "info" is true, the function is called concurrently, so print log only one time.
	if r.isInstalled == 1 {
		var err error
		cnt, err = httpOperations(r.username+"@"+r.hostname, publicKeys, "add")
		if err != nil {
			fmt.Println("Failed to add public keys to remote server via http(" + err.Error() + "), try sftp.")
			cnt = sftpAdd(publicKeys, r, info)
		}
	} else {
		cnt = sftpAdd(publicKeys, r, info)
	}
	if info {
		if cnt <= 1 {
			fmt.Printf("Successfully copied %d key.\n", cnt)
		} else {
			fmt.Printf("Successfully copied %d keys.\n", cnt)
		}
		for _, e := range name {
			insertLink(r.id, e)
		}
		succeed := len(name)
		if succeed <= 1 {
			fmt.Printf("Successfully authorized user %s to log in server %v.\n", name[0], r)
		} else {
			fmt.Printf("Successfully authorized %d users to log in server %v.\n", succeed, r)
		}
	} else {
		if cnt <= 1 {
			fmt.Printf("Successfully copied %d key to %v.\n", cnt, r)
		} else {
			fmt.Printf("Successfully copied %d keys to %v.\n", cnt, r)
		}
	}
}

// copyPublicKeys parse destination and copy public keys to remote server via copyPublicKeysWithRemote.
func copyPublicKeys(destination string, port int, name []string, usePassword bool, privateKey string) {
	r, err := findServer(destination)
	if err != nil {
		r = parseRemote(port, destination, usePassword, privateKey)
	} else if r.password == "" && usePassword {
		r = parseRemote(r.port, destination, true, "")
	} else if privateKey != "" {
		r = parseRemote(r.port, destination, false, privateKey)
	}
	if port != 0 && port != r.port {
		r.port = port
	}
	copyPublicKeysWithRemote(r, name, true)
}

// extractName extracts username from links.
func extractName(links []Link) []string {
	names := make([]string, len(links))
	for i, e := range links {
		names[i] = e.username
	}
	return names
}

// syncPublicKeys syncs public keys to remote server via copyPublicKeysWithRemote.
func syncPublicKeys(serversName []string) {
	servers := make([]Remote, 0)
	if len(serversName) == 0 {
		servers = getServers()
	} else {
		servers = make([]Remote, len(serversName))
		for i, e := range serversName {
			r, err := findServer(e)
			fatalErr(err)
			servers[i] = r
		}
	}
	cnt := len(servers)
	cnt2 := cnt
	// maximum concurrency: 10
	next := make(chan int, 10)
	stop := make(chan int)
	for _, r := range servers {
		// wait to execution
		next <- 1
		go func(des Remote) {
			// sqlite support query concurrently, so query links in coroutines
			copyPublicKeysWithRemote(des, extractName(findLinks(des.id)), false)
			<-next
			// continue main coroutines when all jobs are finished
			cnt -= 1
			if cnt <= 0 {
				stop <- 1
			}
		}(r)
	}
	// wait for jobs are finished
	<-stop
	if cnt2 <= 1 {
		fmt.Printf("Successfully synchronized %d server.\n", cnt2)
	} else {
		fmt.Printf("Successfully synchronized %d servers.\n", cnt2)
	}
}

// showLinks shows links of remote server.
func showLinks() {
	table, _ := gotable.Create("ID", "User", "Server")
	links := getLinks()
	for k, v := range links {
		for _, e := range v {
			_ = table.AddRow([]string{strconv.Itoa(e.id), k, e.String()})
		}
	}

	fmt.Print(table)
}

// delLinks deletes links of remote server via http or sftp (http first).
func delLinks(ids []Link) {
	servers := make(map[int][]Link)
	for _, e := range ids {
		servers[e.serverID] = append(servers[e.serverID], e)
	}
	for k, v := range servers {
		server := findServerById(k)
		deleted := 0
		if server.isInstalled == 1 {
			publicKeys := make([]string, 0)
			for _, e := range v {
				relatedPublicKeys := relatedPublicKey(e)
				for _, h := range relatedPublicKeys {
					if ifRemovePublicKey(e, h) {
						publicKeys = append(publicKeys, h)
					}
				}
			}
			var err error
			deleted, err = httpOperations(server.username+"@"+server.hostname, publicKeys, "del")
			if err != nil {
				fmt.Println("Failed to delete public keys from remote server via http(" + err.Error() + "), try sftp.")
				deleted = sftpDelLinks(v, server)
			} else {
				for _, e := range v {
					deleteLinkById(e.id)
				}
			}
		} else {
			deleted = sftpDelLinks(v, server)
		}
		if deleted == 1 {
			fmt.Printf("Deleted %d key from server %v.\n", deleted, server.String())
		} else {
			fmt.Printf("Deleted %d keys from server %v.\n", deleted, server.String())
		}
	}
}

// delLinksByStrings deletes links by link id via http or sftp (http first).
func delLinksByStrings(ids []string) {
	idi := make([]Link, len(ids))
	for i, e := range ids {
		k, err := strconv.Atoi(e)
		fatalErr(err)
		idi[i] = findLinkById(k)
	}
	delLinks(idi)
}

// sftpDelLinks deletes links via sftp.
func sftpDelLinks(links []Link, r Remote) (deleted int) {
	sftpClient := connect(r, true)
	content := readRemote(sftpClient, r)
	publicKeys := map[string]bool{}
	for _, e := range strings.Split(content, "\n") {
		if e != "" {
			publicKeys[e] = true
		}
	}
	for _, link := range links {
		linkPublicKeys := relatedPublicKey(link)
		for _, e := range linkPublicKeys {
			if ifRemovePublicKey(link, e) {
				publicKeys[e] = false
				deleted += 1
			}
		}
		deleteLinkById(link.id)
	}
	newPublicKeys := make([]string, 0)
	for k, v := range publicKeys {
		if v {
			newPublicKeys = append(newPublicKeys, k)
		}
	}
	newContent := strings.Join(newPublicKeys, "\n")
	writeRemote(sftpClient, getAuthorizedKeysPath(r.username), []byte(newContent))
	return
}
