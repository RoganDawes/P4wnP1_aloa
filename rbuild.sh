#!/bin/bash
#env GOBIN=$(pwd)/build GOOS=linux GOARCH=arm go install cmd/... # compile all main packages to the build folder
env GOOS=linux GOARCH=arm go build -o build/P4wnP1_service cmd/P4wnP1_service/P4wnP1_service.go
env GOOS=linux GOARCH=arm go build -o build/P4wnP1_cli cmd/P4wnP1_cli/P4wnP1_cli.go
scp build/P4wnP1_service pi@raspberrypi.local:~/P4wnP1_go/build
scp build/P4wnP1_cli pi@raspberrypi.local:~/P4wnP1_go/build

#scp P4wnP1_service pi@raspberrypi.local:~/P4wnP1_go
#env GOOS=linux GOARCH=arm go build P4wnP1_cli.go
#scp P4wnP1_cli pi@raspberrypi.local:~/P4wnP1_go

