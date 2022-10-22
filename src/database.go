package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const (
	dbDriverName = "sqlite3"
	dbName       = ".ssh-auth.sqlite"
)

var db *sql.DB

func initDatabase() {
	homeDir, err := os.UserHomeDir()
	checkErr(err)
	db, err = sql.Open(dbDriverName, homeDir+"/"+dbName)
	createTable()
}

func createTable() {
	_, err := db.Exec(`create table if not exists "users" (
		"id" integer primary key autoincrement,
		"username" text not null
	)`)
	checkErr(err)
	_, err = db.Exec(`create table if not exists "servers" (
		"id" integer primary key autoincrement,
		"name" text,
		"hostname" text not null,
		"username" text not null,
		"port" integer not null,
		"password" text not null,
		"publicKeys" text not null
	)`)
	checkErr(err)
	_, err = db.Exec(`create table if not exists "publicKeys" (
		"id" integer primary key autoincrement,
		"publicKeys" text not null,
		"uid" integer,
		foreign key(uid) references users(id)
	)`)
	checkErr(err)
}

func insertUser(username string) bool {
	stmt, err := db.Prepare(`select * from users where username==?`)
	checkErr(err)
	defer checkErr(stmt.Close())
	rows, err := stmt.Query(username)
	checkErr(err)
	exists := rows.Next()
	checkErr(rows.Close())
	if !exists {
		stmt2, err := db.Prepare(`insert into users (username) values(?)`)
		checkErr(err)
		defer checkErr(stmt2.Close())
		_, err = stmt2.Exec(username)
		checkErr(err)
		return true
	}
	return false
}

func insertServer(hostname string, port int, username string, servername string, password string, keys string) bool {
	stmt, err := db.Prepare(`select * from servers where hostname==? and username==? and port==?`)
	checkErr(err)
	defer checkErr(stmt.Close())
	rows, err := stmt.Query(hostname, username, port)
	checkErr(err)
	exists := rows.Next()
	checkErr(rows.Close())
	if !exists {
		stmt2, err := db.Prepare(`insert into servers (name, hostname, username, port, password, publicKeys) values(?, ?, ?, ?, ?, ?)`)
		checkErr(err)
		defer checkErr(stmt2.Close())
		_, err = stmt2.Exec(servername, hostname, username, port, password, keys)
		checkErr(err)
		return true
	}
	stmt2, err := db.Prepare(`update servers set name=?, password=?, publicKeys=? where hostname==? and username==? and port==?`)
	checkErr(err)
	defer checkErr(stmt2.Close())
	_, err = stmt2.Exec(servername, password, keys, hostname, username, port)
	checkErr(err)
	return false
}

func insertPublicKeys(username string, key string) bool {
	stmt, err := db.Prepare(`select * from publicKeys where publicKeys==?`)
	checkErr(err)
	defer checkErr(stmt.Close())
	rows, err := stmt.Query(key)
	checkErr(err)
	exists := rows.Next()
	checkErr(rows.Close())
	if !exists {
		uid := findUserId(username)
		if uid < 0 {
			fmt.Printf("Can't find user %s.\n", username)
		}
		stmt2, err := db.Prepare(`insert into publicKeys (uid, publicKeys) values(?, ?)`)
		checkErr(err)
		defer checkErr(stmt2.Close())
		_, err = stmt2.Exec(uid, key)
		checkErr(err)
		return true
	}
	return false
}

func findUserId(username string) int {
	stmt, err := db.Prepare(`select * from users where username==?`)
	checkErr(err)
	defer checkErr(stmt.Close())
	rows, err := stmt.Query(username)
	checkErr(err)
	defer checkErr(rows.Close())
	if rows.Next() {
		var name string
		var uid int
		checkErr(rows.Scan(&uid, &name))
		return uid
	}
	return -1
}

func findPublicKeys(username string) []string {
	stmt, err := db.Prepare(`select * from users, publicKeys where publicKeys.uid==users.id and username==?`)
	checkErr(err)
	defer checkErr(stmt.Close())
	rows, err := stmt.Query(username)
	checkErr(err)
	defer checkErr(rows.Close())
	result := make([]string, 0)
	for rows.Next() {
		var username, publicKeys string
		var uid, pid, userid int
		checkErr(rows.Scan(&uid, &username, &pid, &publicKeys, &userid))
		result = append(result, publicKeys)
	}
	return result
}

func checkErr(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}
