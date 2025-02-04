@echo off
echo building tool
go build -ldflags="-s -w" -o autonum.exe cmd/main.go
