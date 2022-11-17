ssh-auth: ssh-auth-server src/main.go src/crypto.go src/database.go src/publickeys.go src/server.go src/ssh.go src/users.go
	cp -r bin src
	cd src && go build -o ../ssh-auth

ssh-auth-server: server/main.go server/crypto.go server/database.go server/server.go server/service.go server/watcher.go
	mkdir -p bin
	# cd server/ && CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o "../bin/ssh-auth-server-linux-386"
	cd server/ && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "../bin/ssh-auth-server-linux-amd64"
	# cd server/ && CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o "../bin/ssh-auth-server-linux-arm"
	# cd server/ && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "../bin/ssh-auth-server-linux-arm64"



clean:
	rm ssh-auth
