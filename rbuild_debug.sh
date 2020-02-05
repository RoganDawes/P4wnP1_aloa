#!/bin/bash
echo compiling ...
env GOOS=linux GOARCH=arm GOARM=6 go build -gcflags "all=-N -l" -o build/P4wnP1_service cmd/P4wnP1_service/P4wnP1_service.go
env GOOS=linux GOARCH=arm GOARM=6 go build -o build/P4wnP1_cli cmd/P4wnP1_cli/P4wnP1_cli.go
env GOOS=linux GOARCH=arm GOARM=6 go build -o /tmp/ntest ntest.go

echo uploading ...
scp /tmp/ntest 172.16.0.1:~/P4wnP1/build
scp build/P4wnP1_service 172.16.0.1:~/P4wnP1/build
scp build/P4wnP1_cli 172.16.0.1:~/P4wnP1/build

