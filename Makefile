ssh-auth: src/main.go
	cd src && go mod tidy && go build -o ../ssh-auth

clean:
	rm ssh-auth
