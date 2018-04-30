#!/bin/bash

wget https://storage.googleapis.com/golang/go1.9.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf go1.9.linux-armv6l.tar.gz
export PATH=$PATH:/usr/local/go/bin # put into ~/.profile
echo export PATH=$PATH:/usr/local/go/bin >> ~/.profile
sudo bash -c 'echo export PATH=\$PATH:/usr/local/go/bin >> ~/.profile'
go get google.golang.org/grpc
