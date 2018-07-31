#!/bin/bash

# dependencies for the web app
gopherjs build -o ../build/webapp.js #main.go
#scp ../build/webapp* pi@raspberrypi.local:/usr/local/P4wnP1/www/
#scp ../dist/www/index.html pi@raspberrypi.local:/usr/local/P4wnP1/www/
#scp ../dist/www/p4wnp1.css pi@raspberrypi.local:/usr/local/P4wnP1/www
scp ../build/webapp* root@raspberrypi.local:/usr/local/P4wnP1/www/
scp ../dist/www/index.html root@raspberrypi.local:/usr/local/P4wnP1/www/
scp ../dist/www/p4wnp1.css root@raspberrypi.local:/usr/local/P4wnP1/www/
