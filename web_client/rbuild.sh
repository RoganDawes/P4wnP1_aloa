#!/bin/bash

# dependencies for the web app
gopherjs build -o ../www/webapp.js #main.go
scp ../www/webapp* pi@raspberrypi.local:~/P4wnP1_go/www/
scp ../www/index.html pi@raspberrypi.local:~/P4wnP1_go/www/
scp ../www/p4wnp1.css pi@raspberrypi.local:~/P4wnP1_go/www/
