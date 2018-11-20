#!/bin/bash
echo "$TRIGGER"
if [ "$TRIGGER" = "TRIGGER_USB_GADGET_CONNECTED" ]; then
	echo "USB gadget connected, starting terminal for ttyGS0"
	# start systemd service unit, which spwans agetty for ttyGS0 if needed
	service serial-getty@ttyGS0 start
fi

if [ "$TRIGGER" = "TRIGGER_USB_GADGET_DISCONNECTED" ]; then
	echo "USB gadget disconnected, stopping terminal for ttyGS0"
	# stop systemd service unit, which spwans agetty for ttyGS0
	service serial-getty@ttyGS0 stop
fi
