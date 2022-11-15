package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type OperatorData struct {
	Hash      string `json:"hash"`
	PublicKey string `json:"publicKey"`
	Sign      string `json:"sign"`
}

type ClientData struct {
	Hash      string `json:"hash"`
	PublicKey string `json:"publicKey"`
}

type OperatorResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func getAuthorizedKeysPath(username string) string {
	if username == "root" {
		return "/root/.ssh/authorized_keys"
	}
	return "/home/" + username + "/.ssh/authorized_keys"
}

func addAuthorizedKeys(publicKey, username string) error {
	// Write to /root/.ssh/authorized_keys or /home/username/.ssh/authorized_keys
	path := getAuthorizedKeysPath(username)
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Splitlines and check if the publicKey is already in the file
	linesB := strings.Split(string(data), "\n")
	lines := make([]string, 0)
	for _, line := range linesB {
		if line == publicKey {
			return nil
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	lines = append(lines, strings.TrimSpace(publicKey))
	// Write the file
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}

func delAuthorizedKeys(publicKey, username string) error {
	path := getAuthorizedKeysPath(username)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	linesB := strings.Split(string(data), "\n")
	lines := make([]string, 0)
	for _, line := range linesB {
		if line == publicKey {
			continue
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}

func printLog(message string, err error, r *http.Request) {
	// print ip, method, path, message, error
	if err != nil {
		log.Printf("[%s %s %s] %s: %s\n", r.RemoteAddr, r.Method, r.URL.Path, message, err.Error())
	} else {
		log.Printf("[%s %s %s] %s\n", r.RemoteAddr, r.Method, r.URL.Path, message)
	}
}

func HandleFunc(w http.ResponseWriter, r *http.Request, operator string) {
	if r.Method == "POST" {
		// Parse the request body
		var data OperatorData
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			httpRequestFailed("Error(parse request body)", err, w, r)
			return
		}
		// Get the publicKey and username of client
		publicKey, username, err := getClientPublicKey(data.Hash)
		if err != nil {
			httpRequestFailed("Error(client not found)", err, w, r)
			return
		}
		// Verify the sign
		ok, err := verify(data.PublicKey, publicKey, data.Sign)
		if err != nil {
			httpRequestFailed("Error(verify)", err, w, r)
			return
		}
		if !ok {
			httpRequestFailed("Sign verification failed", nil, w, r)
			return
		}
		if operator == "add" {
			// Add the SSH PublicKey to database
			err = insertSSHPublicKey(data.Hash, data.PublicKey)
		} else if operator == "del" {
			// Delete the SSH PublicKey from database
			err = deleteSSHPublicKey(data.Hash, data.PublicKey)
		}
		if err != nil {
			httpRequestFailed("Error("+operator+":database)", err, w, r)
			return
		}
		if operator == "add" {
			// Add the SSH PublicKey to authorized_keys
			err = addAuthorizedKeys(data.PublicKey, username)
		} else if operator == "del" {
			// Delete the SSH PublicKey from authorized_keys
			err = delAuthorizedKeys(data.PublicKey, username)
		}
		if err != nil {
			httpRequestFailed("Error("+operator+":authorized keys)", err, w, r)
			return
		}
		res := OperatorResult{
			Success: true,
			Message: "success",
		}
		json.NewEncoder(w).Encode(res)
		printLog("Success("+operator+")", nil, r)
	} else {
		httpRequestFailed("Error(method)", nil, w, r)
	}
}

func server() {
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		HandleFunc(w, r, "add")
	})

	http.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
		HandleFunc(w, r, "del")
	})
	log.Println("Server started at :22222")
	log.Fatal(http.ListenAndServe(":22222", nil))
}

func httpRequestFailed(info string, err error, w http.ResponseWriter, r *http.Request) {
	res := OperatorResult{
		Success: false,
		Message: info,
	}
	if err != nil {
		res.Message += ": " + err.Error()
	}
	json.NewEncoder(w).Encode(res)
	printLog(info, err, r)
}
