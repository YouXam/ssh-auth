package main

const serviceFile = `[Unit]
Description=ssh-auth server
After=network.target
ConditionPathExists=/usr/local/bin/ssh-auth-server

[Service]
ExecStart=/usr/local/bin/ssh-auth-server`
