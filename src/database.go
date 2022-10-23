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
	fatalErr(err)
	db, err = sql.Open(dbDriverName, homeDir+"/"+dbName)
	createTable()
}

func createTable() {
	_, err := db.Exec(`create table if not exists "users" (
		"id" integer primary key autoincrement,
		"username" text not null
	)`)
	fatalErr(err)
	_, err = db.Exec(`create table if not exists "servers" (
		"id" integer primary key autoincrement,
		"name" text,
		"hostname" text not null,
		"username" text not null,
		"port" integer not null,
		"password" text not null,
		"privateKey" text not null
	)`)
	fatalErr(err)
	_, err = db.Exec(`create table if not exists "publicKeys" (
		"id" integer primary key autoincrement,
		"publicKey" text not null,
		"username" text not null
	)`)
	fatalErr(err)
	_, err = db.Exec(`create table if not exists "links" (
		"id" integer primary key autoincrement,
		"username" text not null,
		"serverID" integer not null
	)`)
	fatalErr(err)
}

func insertUser(username string) bool {
	stmt, err := db.Prepare(`select * from users where username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(username)
	fatalErr(err)
	exists := rows.Next()
	fatalErr(rows.Close())
	if !exists {
		stmt2, err := db.Prepare(`insert into users (username) values(?)`)
		fatalErr(err)
		defer func() { fatalErr(stmt2.Close()) }()
		_, err = stmt2.Exec(username)
		fatalErr(err)
		return true
	}
	return false
}
func insertLink(serverId int, username string) bool {
	stmt, err := db.Prepare(`select * from links where username==? and serverID=?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(username, serverId)
	fatalErr(err)
	exists := rows.Next()
	fatalErr(rows.Close())
	if !exists {
		stmt2, err := db.Prepare(`insert into links (username, serverID) values(?, ?)`)
		fatalErr(err)
		defer func() { fatalErr(stmt2.Close()) }()
		_, err = stmt2.Exec(username, serverId)
		fatalErr(err)
		return true
	}
	return false
}

func findLinks(serverId int) []string {
	stmt, err := db.Prepare(`select * from links where serverID==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(serverId)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := make([]string, 0)
	for rows.Next() {
		var username string
		var id, serverID int
		fatalErr(rows.Scan(&id, &username, &serverID))
		result = append(result, username)
	}
	return result
}

func findServerByName(servername string) (Remote, error) {
	stmt, err := db.Prepare(`select * from servers where name==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(servername)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	if rows.Next() {
		var name, hostname, username, password, privateKey string
		var port, id int
		fatalErr(rows.Scan(&id, &name, &hostname, &username, &port, &password, &privateKey))
		return Remote{
			username,
			hostname,
			port,
			password,
			privateKey,
			name,
			id,
		}, nil
	}
	return Remote{}, fmt.Errorf("can not find server %s", servername)
}

func insertServer(hostname string, port int, username string, servername string, password string, key string) bool {
	pre, err := findServerByName(servername)
	if err == nil && (pre.hostname != hostname || pre.username != username || pre.password != password) {
		fatalErr(fmt.Errorf("server %s has already exists", hostname))
	}
	stmt, err := db.Prepare(`select * from servers where hostname==? and username==? and port==?`)
	fatalErr(err)
	defer func() {
		fatalErr(stmt.Close())
	}()
	rows, err := stmt.Query(hostname, username, port)
	fatalErr(err)
	exists := rows.Next()
	fatalErr(rows.Close())
	if !exists {
		stmt2, err := db.Prepare(`insert into servers (name, hostname, username, port, password, privateKey) values(?, ?, ?, ?, ?, ?)`)
		fatalErr(err)
		defer func() { fatalErr(stmt.Close()) }()
		_, err = stmt2.Exec(servername, hostname, username, port, password, key)
		fatalErr(err)
		return true
	} else {
		stmt2, err := db.Prepare(`update servers set name=?, password=?, privateKey=? where hostname==? and username==? and port==?`)
		fatalErr(err)
		defer func() { fatalErr(stmt2.Close()) }()
		_, err = stmt2.Exec(servername, password, key, hostname, username, port)
		fatalErr(err)
		return false
	}
}

func findServer(destination string) (Remote, error) {
	stmt, err := db.Prepare(`select * from servers where name==? or username||'@'||hostname==? or hostname==? and username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	home, err := os.UserHomeDir()
	fatalErr(err)
	rows, err := stmt.Query(destination, destination, destination, home)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	if rows.Next() {
		var name, hostname, username, password, privateKey string
		var port, id int
		fatalErr(rows.Scan(&id, &name, &hostname, &username, &port, &password, &privateKey))
		return Remote{
			username,
			hostname,
			port,
			password,
			privateKey,
			name,
			id,
		}, nil
	}
	return Remote{}, fmt.Errorf("can not find server %s", destination)
}

func insertPublicKeys(username string, key string) bool {
	stmt, err := db.Prepare(`select * from publicKeys where publicKey==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(key)
	fatalErr(err)
	exists := rows.Next()
	fatalErr(rows.Close())
	if !exists {
		if findUserId(username) < 0 {
			fmt.Printf("Can't find user %s.\nExited.\n", username)
			os.Exit(1)
		}
		stmt2, err := db.Prepare(`insert into publicKeys (username, publicKey) values(?, ?)`)
		fatalErr(err)
		defer func() { fatalErr(stmt2.Close()) }()
		_, err = stmt2.Exec(username, key)
		fatalErr(err)
		return true
	}
	return false
}

func findUserId(username string) int {
	stmt, err := db.Prepare(`select * from users where username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(username)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	if rows.Next() {
		var name string
		var uid int
		fatalErr(rows.Scan(&uid, &name))
		return uid
	}
	return -1
}

func getUsers() []string {
	stmt, err := db.Prepare(`select * from users`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query()
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := make([]string, 0)
	for rows.Next() {
		var username string
		var id int
		fatalErr(rows.Scan(&id, &username))
		result = append(result, username)
	}
	return result
}

func getServers() []Remote {
	stmt, err := db.Prepare(`select * from servers`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query()
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := make([]Remote, 0)
	for rows.Next() {
		var name, hostname, username, password, privateKey string
		var port, id int
		fatalErr(rows.Scan(&id, &name, &hostname, &username, &port, &password, &privateKey))
		result = append(result, Remote{
			username,
			hostname,
			port,
			password,
			privateKey,
			name,
			id,
		})
	}
	return result
}

func findPublicKeys(username string) []string {
	stmt, err := db.Prepare(`select * from users, publicKeys where publicKeys.username==users.username and users.username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(username)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := make([]string, 0)
	for rows.Next() {
		var username, publicKeys, username2 string
		var uid, pid int
		fatalErr(rows.Scan(&uid, &username, &pid, &publicKeys, &username2))
		result = append(result, publicKeys)
	}
	return result
}

func fatalErr(e error) {
	if e != nil {
		fmt.Println(e)
		fmt.Println("Exited.")
		os.Exit(1)
	}
}

func fatalErrRemote(r Remote, e error) {
	if e != nil {
		fmt.Printf("%s: %v\n", r.toString(), e)
		fmt.Println("Exited.")
		os.Exit(1)
	}
}
