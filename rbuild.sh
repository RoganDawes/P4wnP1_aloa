#!/bin/bash
env GOOS=linux GOARCH=arm go build P4wnP1_service.go
scp P4wnP1_service pi@raspberrypi.local:~/P4wnP1_go
env GOOS=linux GOARCH=arm go build P4wnP1_cli.go
scp P4wnP1_cli pi@raspberrypi.local:~/P4wnP1_go

