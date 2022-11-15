package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbDriverName = "sqlite3"
	dbName       = "ssh-auth-server.sqlite"
)

var db *sql.DB

func initDatabase() {
	var err error
	// check if the directory exists
	if _, err = os.Stat("/var/lib/ssh-auth-server"); os.IsNotExist(err) {
		runCommnd("mkdir /var/lib/ssh-auth-server", "sudo", "mkdir", "-p", "/var/lib/ssh-auth-server")
		// assign permission to everyone
		runCommnd("chmod 777 /var/lib/ssh-auth-server", "sudo", "chmod", "777", "/var/lib/ssh-auth-server")
	}
	db, err = sql.Open(dbDriverName, "/var/lib/ssh-auth-server/"+dbName)
	fatalErr(err)
	createTable()
}

func createTable() {
	_, err := db.Exec(`create table if not exists "clientPublicKeys" (
		"id" integer primary key autoincrement,
		"hash" text not null,
		"username" text not null,
		"publicKey" text not null
	)`)
	fatalErr(err)
	_, err = db.Exec(`create table if not exists "sshPublicKeys" (
		"id" integer primary key autoincrement,
		"hash" text not null,
		"publicKey" text not null
	)`)
	fatalErr(err)
	_, err = db.Exec(`create table if not exists "usernames" (
		"id" integer primary key autoincrement,
		"username" text not null
	)`)
	fatalErr(err)
}

func insertClientPublicKey(hash string, publicKey string) error {
	rows, err := db.Query("select id from clientPublicKeys where hash = ?", hash)
	if err != nil {
		return err
	}
	if rows.Next() {
		rows.Close()
		return errors.New("hash already exists")
	}
	rows.Close()
	username := os.Getenv("USER")
	insertUsername(username)
	_, err = db.Exec("insert into clientPublicKeys (hash, username, publicKey) values (?, ?, ?)", hash, username, publicKey)
	return err
}

func getClientPublicKey(hash string) (string, string, error) {
	rows, err := db.Query("select publicKey,username from clientPublicKeys where hash = ?", hash)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()
	if rows.Next() {
		var publicKey, username string
		err = rows.Scan(&publicKey, &username)
		if err != nil {
			return "", "", err
		}
		return publicKey, username, nil
	}
	return "", "", errors.New("client not found")
}

func insertSSHPublicKey(hash string, publicKey string) error {
	_, err := db.Exec("insert into sshPublicKeys (hash, publicKey) values (?, ?)", hash, publicKey)
	return err
}

func deleteSSHPublicKey(hash string, publicKey string) error {
	_, err := db.Exec("delete from sshPublicKeys where hash = ? and publicKey = ?", hash, publicKey)
	return err
}

func insertUsername(username string) error {
	_, err := db.Exec("insert into usernames (username) select ? where not exists (select * from usernames where username = ?)", username, username)
	return err
}

func getSSHPublicKeysByUsername(username string) ([]string, error) {
	rows, err := db.Query("select publicKey from sshPublicKeys where hash in (select hash from clientPublicKeys where username = ?)", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var publicKeys []string
	for rows.Next() {
		var publicKey string
		err = rows.Scan(&publicKey)
		if err != nil {
			return nil, err
		}
		publicKeys = append(publicKeys, publicKey)
	}
	return publicKeys, nil
}

func getUsernames() []string {
	rows, err := db.Query("select username from usernames")
	fatalErr(err)
	defer rows.Close()
	var usernames []string
	for rows.Next() {
		var username string
		err = rows.Scan(&username)
		fatalErr(err)
		usernames = append(usernames, username)
	}
	return usernames
}

func fatalErr(e error) {
	if e != nil {
		log.Println(e)
		log.Println("Exited.")
		os.Exit(1)
	}
}
