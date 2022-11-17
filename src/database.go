package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
)

const (
	dbDriverName = "sqlite"
	dbName       = ".ssh-auth.sqlite"
)

type Link struct {
	id       int
	username string
	serverID int
}

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
		"privateKey" text not null,
		"isInstalled" int not null
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
	_, err = db.Exec(`create table if not exists "keyPair"(
    	"id" integer primary key autoincrement,
    	"destination" text not null,
    	"publicKey" text not null,
    	"privateKey" text not null
	)`)
	fatalErr(err)
}

func getKeyPair(destination string) (string, string) {
	stmt, err := db.Prepare(`select * from keyPair where destination==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(destination)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	if rows.Next() {
		var id int
		var publicKey, privateKey, destination string
		fatalErr(rows.Scan(&id, &destination, &publicKey, &privateKey))
		return privateKey, publicKey
	}
	return "", ""
}

func insertKeyPair(destination, privateKey, publicKey string) {
	stmt, err := db.Prepare(`insert into keyPair (destination, privateKey, publicKey) values(?, ?, ?)`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	_, err = stmt.Exec(destination, privateKey, publicKey)
	fatalErr(err)
}

func insertUser(username string) bool {
	// check weather the user has already exists
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

func deleteUser(username string) {
	stmt, err := db.Prepare(`delete from users where username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	_, err = stmt.Exec(username)
	fatalErr(err)
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

func insertLink(serverId int, username string) bool {
	// check weather the link has already exists
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

func findLinks(serverId int) []Link {
	stmt, err := db.Prepare(`select * from links where serverID==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(serverId)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := make([]Link, 0)
	for rows.Next() {
		var username string
		var id, serverID int
		fatalErr(rows.Scan(&id, &username, &serverID))
		result = append(result, Link{
			id,
			username,
			serverID,
		})
	}
	return result
}

func findLinksByUsername(username string) []Link {
	stmt, err := db.Prepare(`select * from links where username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(username)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := make([]Link, 0)
	for rows.Next() {
		var username string
		var id, serverID int
		fatalErr(rows.Scan(&id, &username, &serverID))
		result = append(result, Link{
			id,
			username,
			serverID,
		})
	}
	return result
}

func findLinkById(id int) Link {
	stmt, err := db.Prepare(`select * from links where id==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(id)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	if rows.Next() {
		var username string
		var id, serverID int
		fatalErr(rows.Scan(&id, &username, &serverID))
		return Link{
			id:       id,
			username: username,
			serverID: serverID,
		}
	}
	fatalErr(fmt.Errorf("can not find link #%d", id))
	return Link{}
}

func getLinks() map[string][]Remote {
	stmt, err := db.Prepare(`select links.id, links.username, servers.name, servers.username, servers.hostname, servers.port from links, servers where links.serverID==servers.id`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query()
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := map[string][]Remote{}
	for rows.Next() {
		var linkUsername, serverName, serverUsername, hostname string
		var id, port int
		fatalErr(rows.Scan(&id, &linkUsername, &serverName, &serverUsername, &hostname, &port))
		result[linkUsername] = append(result[linkUsername], Remote{
			username:   serverUsername,
			hostname:   hostname,
			port:       port,
			servername: serverName,
			id:         id,
		})
	}
	return result
}

func deleteLinkById(id int) {
	stmt, err := db.Prepare(`delete from links where id==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	_, err = stmt.Exec(id)
	fatalErr(err)
}

func deleteLinkByServerID(id int) {
	stmt, err := db.Prepare(`delete from links where serverID==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	_, err = stmt.Exec(id)
	fatalErr(err)
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
		var port, id, installedDaemon int
		fatalErr(rows.Scan(&id, &name, &hostname, &username, &port, &password, &privateKey, &installedDaemon))
		return Remote{
			username,
			hostname,
			port,
			password,
			privateKey,
			name,
			id,
			installedDaemon,
		}, nil
	}
	return Remote{}, fmt.Errorf("can not find server %s", servername)
}

func findServerById(id int) Remote {
	stmt, err := db.Prepare(`select * from servers where id==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(id)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	if rows.Next() {
		var name, hostname, username, password, privateKey string
		var port, id, installedDaemon int
		fatalErr(rows.Scan(&id, &name, &hostname, &username, &port, &password, &privateKey, &installedDaemon))
		return Remote{
			username,
			hostname,
			port,
			password,
			privateKey,
			name,
			id,
			installedDaemon,
		}
	}
	fatalErr(fmt.Errorf("can not find server #%d", id))
	return Remote{}
}

func deleteServerByID(id int) {
	stmt, err := db.Prepare(`delete from servers where id==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	_, err = stmt.Exec(id)
	fatalErr(err)
}

func insertServer(hostname string, port int, username string, servername string, password string, key string, isInstalled int) bool {
	pre, err := findServerByName(servername)
	if err == nil && (pre.hostname == hostname && pre.username == username && pre.port == port) {
		return false
	}
	if err == nil && (pre.hostname != hostname || pre.username != username || pre.port != port) {
		fatalErr(fmt.Errorf("server %v has already exists", servername))
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
		stmt2, err := db.Prepare(`insert into servers (name, hostname, username, port, password, privateKey, isInstalled) values(?, ?, ?, ?, ?, ?, ?)`)
		fatalErr(err)
		defer func() { fatalErr(stmt.Close()) }()
		_, err = stmt2.Exec(servername, hostname, username, port, password, key, isInstalled)
		fatalErr(err)
		return true
	} else {
		stmt2, err := db.Prepare(`update servers set name=?, password=?, privateKey=?, isInstalled=? where hostname==? and username==? and port==?`)
		fatalErr(err)
		defer func() { fatalErr(stmt2.Close()) }()
		_, err = stmt2.Exec(servername, password, key, isInstalled, hostname, username, port)
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
		var port, id, installedDaemon int
		fatalErr(rows.Scan(&id, &name, &hostname, &username, &port, &password, &privateKey, &installedDaemon))
		return Remote{
			username,
			hostname,
			port,
			password,
			privateKey,
			name,
			id,
			installedDaemon,
		}, nil
	}
	return Remote{}, fmt.Errorf("can not find server %s", destination)
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
		var port, id, installedDaemon int
		fatalErr(rows.Scan(&id, &name, &hostname, &username, &port, &password, &privateKey, &installedDaemon))
		result = append(result, Remote{
			username,
			hostname,
			port,
			password,
			privateKey,
			name,
			id,
			installedDaemon,
		})
	}
	return result
}

func insertPublicKeys(username string, key string) bool {
	stmt, err := db.Prepare(`select * from publicKeys where publicKey==? and username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(key, username)
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

func relatedPublicKey(link Link) []string {
	stmt, err := db.Prepare(`select publicKey from links, publicKeys, servers where publicKeys.username==links.username and servers.id==links.serverID and servers.id == ? and links.username==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(link.serverID, link.username)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	result := make([]string, 0)
	for rows.Next() {
		var publicKeys string
		fatalErr(rows.Scan(&publicKeys))
		result = append(result, publicKeys)
	}
	return result
}

func ifRemovePublicKey(link Link, publicKey string) bool {
	stmt, err := db.Prepare(`select * from publicKeys, links where publicKeys.username == links.username and publicKeys.username!=? and serverID==? and publicKey==?`)
	fatalErr(err)
	defer func() { fatalErr(stmt.Close()) }()
	rows, err := stmt.Query(link.username, link.serverID, publicKey)
	fatalErr(err)
	defer func() { fatalErr(rows.Close()) }()
	if rows.Next() {
		return false
	}
	return true
}

func fatalErrRemote(r Remote, e error) {
	if e != nil {
		fmt.Printf("%v: %v\n", r, e)
		fmt.Println("Exited.")
		os.Exit(1)
	}
}

func fatalErr(e error) {
	if e != nil {
		fmt.Println(e)
		fmt.Println("Exited.")
		os.Exit(1)
	}
}
