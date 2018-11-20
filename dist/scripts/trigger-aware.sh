#!/bin/bash
if [ "$TRIGGER" = "" ]; then
	echo "Script not called from TriggerAction"
	exit
fi

if [ "$TRIGGER" = "TRIGGER_SERVICE_STARTED" ]; then
	echo "Script called from service started"
fi

if [ "$TRIGGER" = "TRIGGER_USB_GADGET_CONNECTED" ]; then
	echo "Script called from Trigger USB gadget connected"
fi

if [ "$TRIGGER" = "TRIGGER_USB_GADGET_DISCONNECTED" ]; then
	echo "Script called from Trigger USB gadget disconnected"
fi

if [ "$TRIGGER" = "TRIGGER_WIFI_AP_STARTED" ]; then
	echo "Script called from Trigger WiFi Access Point started"
fi

if [ "$TRIGGER" = "TRIGGER_WIFI_CONNECTED_AS_STA" ]; then
	echo "Script called from Trigger Connected to existing WiFi"
fi

if [ "$TRIGGER" = "TRIGGER_SSH_LOGIN" ]; then
	echo "Script called from Trigger SSH login for user: $SSH_LOGIN_USER"
fi

if [ "$TRIGGER" = "TRIGGER_DHCP_LEASE_GRANTED" ]; then
	echo "Script called from Trigger DHCP lease granted"
	echo "\tInterface: $DHCP_LEASE_IFACE"
	echo "\tMac:       $DHCP_LEASE_MAC"
	echo "\tIP:        $DHCP_LEASE_IP"
	echo "\tHost:      $DHCP_LEASE_HOST"
fi

if [ "$TRIGGER" = "TRIGGER_GROUP_RECEIVE" ]; then
	echo "Script called from Trigger Received value on group channel"
	echo "\tGroup: $GROUP"
	echo "\tValue: $VALUE"
fi

if [ "$TRIGGER" = "TRIGGER_GROUP_RECEIVE_MULTI" ]; then
	echo "Script called from Trigger Received multiple values on group channel"
	echo "\tGroup:  $GROUP"
	echo "\tValues: $VALUES"
	echo "\tType:   $MULTI_TYPE"
fi

if [ "$TRIGGER" = "TRIGGER_GPIO_IN" ]; then
	echo "Script called from Trigger GPIO in"
	echo "\tGPIO pin:  $GPIO_PIN"
	echo "\tPin Level: $GPIO_LEVEL"
fi

