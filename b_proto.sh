#!/bin/bash
# golang version
protoc -I proto/ proto/grpc.proto --go_out=plugins=grpc:proto

# gopherjs version
protoc -I proto/ proto/grpc.proto --gopherjs_out=plugins=grpc:proto/gopherjs

