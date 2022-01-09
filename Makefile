SHELL := /bin/bash
#PATH := /usr/local/go/bin:$(PATH)

all: compile

test:
#	export PATH="$$PATH:/usr/local/go/bin" # put into ~/.profile
	echo $(pwd)
	echo $HOME

# make dep runs without sudo
# stay in $HOME/P4wnP1
# ONLY RUN -------- make dep / make compile / make installkali
dep:
	#Disable GUI Login
	sudo systemctl set-default multi-user.target
	##Disabling ipv6 for me makes repos work faster
	echo -e "net.ipv6.conf.all.disable_ipv6 = 1\nnet.ipv6.conf.default.disable_ipv6 = 1" | sudo tee -a /etc/sysctl.conf
	sudo sysctl -p
	sudo apt-get update
	sudo apt-get -y install git screen hostapd autossh bluez bluez-tools bridge-utils policykit-1 genisoimage iodine haveged tcpdump

	#sudo apt-get -y install python-pip python-dev

	# before installing dnsmasq, the nameserver from /etc/resolv.conf should be saved
	# to restore after install (gets overwritten by dnsmasq package)
	cp /etc/resolv.conf /tmp/backup_resolv.conf
	sudo apt-get -y install dnsmasq
	sudo /bin/bash -c 'cat /tmp/backup_resolv.conf > /etc/resolv.conf'
    	
	sudo apt-get -y install dhcpcd5

	# python dependencies for HIDbackdoor
	sudo pip install pycrypto # already present on stretch
	#sudo pip install pydispatcher #already present

	sudo apt-get -y install golang-go

	# put into ~/.profile
	# ToDo: check if already present
	#echo "export PATH=\$$PATH:/usr/local/go/bin" >> ~/.profile
	#sudo bash -c 'echo export PATH=\$$PATH:/usr/local/go/bin >> ~/.profile'

	# install gopherjs
	go get -u github.com/gopherjs/gopherjs

	# we don't need protoc + protoc-grpc-web, because the proto file is shipped pre-compiled

	# go dependencies for webapp (without my own)
	#go get google.golang.org/grpc
	#go get -u github.com/improbable-eng/grpc-web/go/grpcweb
	#go get -u github.com/gorilla/websocket

compile:
	go get -u github.com/mame82/P4wnP1_aloa/... # partially downloads again, but we need the library packages in go path to build
    	#
	go get -d github.com/mame82/P4wnP1_aloa/...
	# <--- second compilation, maybe -d flag on go get above is better
	env GOBIN=$(pwd)/build go install ./cmd/... # compile all main packages to the build folder

	# compile the web app
	# ToDo: (check if dependencies have been fetched by 'go get', even with the build js tags)
	$HOME/go/bin/gopherjs get github.com/mame82/P4wnP1_aloa/web_client/...
	$HOME/go/bin/gopherjs build -m -o build/webapp.js web_client/*.go

installkali:
	#apt-get -y install git screen hostapd autossh bluez bluez-tools bridge-utils policykit-1 genisoimage iodine haveged
	#apt-get -y install tcpdump
	#apt-get -y install python-pip python-dev

	# before installing dnsmasq, the nameserver from /etc/resolv.conf should be saved
	# to restore after install (gets overwritten by dnsmasq package)
	#cp /etc/resolv.conf /tmp/backup_resolv.conf
	#apt-get -y install dnsmasq
	#/bin/bash -c 'cat /tmp/backup_resolv.conf > /etc/resolv.conf'

	# python dependencies for HIDbackdoor
	#sudo pip install pydispatcher

	cp build/P4wnP1_service /usr/local/bin/
	cp build/P4wnP1_cli /usr/local/bin/
	cp dist/P4wnP1.service /etc/systemd/system/P4wnP1.service
	# copy over keymaps, scripts and www data
	mkdir -p /usr/local/P4wnP1
	cp -R dist/keymaps /usr/local/P4wnP1/
	cp -R dist/scripts /usr/local/P4wnP1/
	cp -R dist/HIDScripts /usr/local/P4wnP1/
	cp -R dist/www /usr/local/P4wnP1/
	cp -R dist/db /usr/local/P4wnP1/
	cp -R dist/helper /usr/local/P4wnP1/
	cp -R dist/ums /usr/local/P4wnP1/
	cp -R dist/legacy /usr/local/P4wnP1/
	#cp build/webapp.js /usr/local/P4wnP1/www
    	#
	cp ./webapp.js /usr/local/P4wnP1/www
	cp build/webapp.js.map /usr/local/P4wnP1/www
	echo "dtoverlay=dwc2" | sudo tee -a /boot/config.txt

	# careful testing
	#sudo update-rc.d dhcpcd disable
	#sudo update-rc.d dnsmasq disable
	systemctl disable networking.service # disable network service, relevant parts are wrapped by P4wnP1 (boottime below 20 seconds)

	# enable service
	systemctl enable haveged
	systemctl enable avahi-daemon
	systemctl enable P4wnP1.service

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
	cp -R dist/db /usr/local/P4wnP1/
	#cp dist/bin/* /usr/local/bin/
	cp build/webapp.js /usr/local/P4wnP1/www
	cp build/webapp.js.map /usr/local/P4wnP1/www

	# careful testing
	#sudo update-rc.d dhcpcd disable
	#sudo update-rc.d dnsmasq disable
	systemctl disable networking.service # disable network service, relevant parts are wrapped by P4wnP1 (boottime below 20 seconds)

	# reinit service daemon
	systemctl daemon-reload
	# enable service
	systemctl enable haveged
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
