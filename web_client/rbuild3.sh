#!/bin/bash

# dependencies for the web app
gopherjs build -o ../build/webapp.js #main.go
#scp ../build/webapp* pi@raspberrypi.local:/usr/local/P4wnP1/www/
#scp ../dist/www/index.html pi@raspberrypi.local:/usr/local/P4wnP1/www/
#scp ../dist/www/p4wnp1.css pi@raspberrypi.local:/usr/local/P4wnP1/www
scp ../build/webapp* 172.16.0.1:/usr/local/P4wnP1/www/
scp ../dist/www/index.html 172.16.0.1:/usr/local/P4wnP1/www/
scp ../dist/www/p4wnp1.css 172.16.0.1:/usr/local/P4wnP1/www/
