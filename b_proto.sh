#!/bin/bash
protoc -I proto/ proto/grpc.proto --go_out=plugins=grpc:proto
