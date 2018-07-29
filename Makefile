all: build

# make dep runs without sudo
dep:
    sudo apt-get -y install git screen hostapd autossh bluez bluez-tools bridge-utils policykit-1 genisoimage iodine haveged
    sudo apt-get -y install tcpdump
    sudo apt-get -y install python-pip python-dev

    # before installing dnsmasq, the nameserver from /etc/resolv.conf should be saved
    # to restore after install (gets overwritten by dnsmasq package)
    cp /etc/resolv.conf /tmp/backup_resolv.conf
    sudo apt-get -y install dnsmasq
    sudo /bin/bash -c 'cat /tmp/backup_resolv.conf > /etc/resolv.conf'

    # python dependencies for HIDbackdoor
    sudo pip install pycrypto # already present on stretch
    sudo pip install pydispatcher

    # install go
    wget https://storage.googleapis.com/golang/go1.9.linux-armv6l.tar.gz
    sudo tar -C /usr/local -xzf go1.9.linux-armv6l.tar.gz
    export PATH=$PATH:/usr/local/go/bin # put into ~/.profile
    echo export PATH=$PATH:/usr/local/go/bin >> ~/.profile
    sudo bash -c 'echo export PATH=\$PATH:/usr/local/go/bin >> ~/.profile'

    # install gopherjs
    go get -u github.com/gopherjs/gopherjs

    # we don't need protoc + protoc-grpc-web, because the proto file is shipped pre-compiled

    # go dependencies for webapp (without my own)
    #go get google.golang.org/grpc
    #go get -u github.com/improbable-eng/grpc-web/go/grpcweb
    #go get -u github.com/gorilla/websocket

build:
    go get -u github.com/mame82/P4wnP1_go # partially downloads again, but we need the library packages in go path to build
	env GOBIN=$(pwd)/build go install ./... # compile all main packages to the build folder

	# compile the web app
	# ToDo: (check if dependencies have been fetched by 'go get', even with the build js tags)
	gopherjs build -m -o build/webapp.js web_client/*.go

install:
	cp build/P4wnP1_service /usr/local/bin/
	cp build/P4wnP1_cli /usr/local/bin/
	cp dist/P4wnP1.service /etc/systemd/system/P4wnP1.service
	# copy over keymaps, scripts and www data
	mkdir -p /usr/local/P4wnP1
	cp -R dist/keymaps /usr/local/P4wnP1/
	cp -R dist/scripts /usr/local/P4wnP1/
	cp -R dist/HIDScripts /usr/local/P4wnP1/
	cp -R dist/www /usr/local/P4wnP1/
	cp build/webapp.js /usr/local/P4wnP1/www
	cp build/webapp.js.map /usr/local/P4wnP1/www

	# careful testing
	#sudo update-rc.d dhcpcd disable
    #sudo update-rc.d dnsmasq disable

	# reinit service daemon
	systemctl daemon-reload
	# enable service
	systemctl enable P4wnP1.service
	# start service
	service P4wnP1 start

remove:
	# stop service
	service P4wnP1 stop
	# disable service
	systemctl disable P4wnP1.service
	rm -f /usr/local/bin/P4wnP1_service
	rm -f /usr/local/bin/P4wnP1_cli
	rm -f /etc/systemd/system/P4wnP1.service
	rm -R /usr/local/P4wnP1/    # this folder should be kept, if only an update should be applied
	# reinit service daemon
	systemctl daemon-reload

	#sudo update-rc.d dhcpcd enable

