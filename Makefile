all: build

build:
	go build P4wnP1_service.go
	go build P4wnP1_cli.go

install:
	cp P4wnP1_service /usr/local/bin/
	cp P4wnP1_cli /usr/local/bin/
	cp P4wnP1.service /etc/systemd/system/P4wnP1.service
	mkdir /usr/local/P4wnP1
	cp -R keymaps /usr/local/P4wnP1/
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
	rm -R /usr/local/P4wnP1/
	# reinit service daemon
	systemctl daemon-reload
