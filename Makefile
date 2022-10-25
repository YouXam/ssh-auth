ssh-auth: src/main.go
	cd src && go build -o ../ssh-auth

clean:
	rm ssh-auth
