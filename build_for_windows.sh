#!/bin/sh
GOOS=windows GOARCH=386 go build -o go_client.exe client.go
