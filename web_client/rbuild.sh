#!/bin/bash

# dependencies for the web app
gopherjs build -o ../build/webapp.js #main.go
scp ../build/webapp* pi@raspberrypi.local:~/P4wnP1_go/dist/www/
scp ../dist/www/index.html pi@raspberrypi.local:~/P4wnP1_go/dist/www/
scp ../dist/www/p4wnp1.css pi@raspberrypi.local:~/P4wnP1_go/dist/www/
