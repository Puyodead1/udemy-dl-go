@echo off

setlocal
set GOOS=linux
set GOARCH=amd64
echo Building...
go build -o ./dist/udemy-dl-go-linux-amd64 -v ./      

endlocal
