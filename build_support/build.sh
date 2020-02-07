#!/bin/bash

# has to be run from 'build_support' subfolder
cd ..
echo "compiling P4wnP1_cli and P4wnP1_service ..."
env GOOS=linux GOARCH=arm GOARM=6 go build -o build/P4wnP1_service cmd/P4wnP1_service/P4wnP1_service.go
env GOOS=linux GOARCH=arm GOARM=6 go build -o build/P4wnP1_cli cmd/P4wnP1_cli/P4wnP1_cli.go

echo "compiling web client to JavaScript ..."
cd web_client
gopherjs build -o ../build/webapp.js

echo "...Results stored in ./build directory"
echo
echo "On P4wnP1 ALOA the compiled files have to be placed at the following"
echo "locations:"
echo
echo "    /usr/local/bin/P4wnP1_cli"
echo "    /usr/local/bin/P4wnP1_service"
echo "    /usr/local/P4wnP1/www/webapp.js"
echo "    /usr/local/P4wnP1/www/webapp.js.map"

