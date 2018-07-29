#!/bin/bash
env GOBIN=$(pwd)/build GOOS=linux GOARCH=arm go install ./... # compile all main packages to the build folder
scp build/P4wnP1_service pi@raspberrypi.local:~/P4wnP1_go
scp build/P4wnP1_cli pi@raspberrypi.local:~/P4wnP1_go

#env GOOS=linux GOARCH=arm go build P4wnP1_service.go
#scp P4wnP1_service pi@raspberrypi.local:~/P4wnP1_go
#env GOOS=linux GOARCH=arm go build P4wnP1_cli.go
#scp P4wnP1_cli pi@raspberrypi.local:~/P4wnP1_go

