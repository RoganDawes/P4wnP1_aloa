#!/bin/bash

# Install gopherjs-gRPC plugin by Johan Brandhorst

# go and protoc have to be installed already
# $GOPATH/bin has to be in path

go get -u github.com/gopherjs/gopherjs
go get -u github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs