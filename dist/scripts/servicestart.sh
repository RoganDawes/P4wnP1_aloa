#!/bin/bash

# Enable USB functions RNDIS, CDC ECM (don't disable other functions which already have been enabled)
P4wnP1_cli USB set --rndis 1 --cdc-ecm 1

# Configure USB ethernet interface "usbeth" to run a DHCP server
#   - use IPv4 172.16.0.1 for interface with netmask 255.255.255.252
#   - disable DHCP option 3 (router) by passing an empty value
#   - disable DHCP option 6 (DNS) by passing an empty value
#   - add a DHCP range from 172.16.0.2 to 172.16.0.2 (single IP) with a lease time of 1 minute
P4wnP1_cli NET set server -i usbeth -a 172.16.0.1 -m 255.255.255.252 -o "3:" -o "6:" -r "172.16.0.2|172.16.0.2|1m"

# Enable WiFi AP (reg US, channel 6, SSID/AP name: "P4wnP1", pre shared key: "MaMe82-P4wnP1", don't use nexmon firmware)
# Note: As a pre-shared key is given, P4wnP1 assume the AP should use WPA2-PSK
# Note 2: The SSID uses Unicode characters not necessarily supported by the console, but P4wnP1 supports UTF-8 ;-)
P4wnP1_cli WIFI set ap -r US -c 6 -s "💥🖥💥 Ⓟ➃ⓌⓃ🅟❶" -k "MaMe82-P4wnP1" --nonexmon

# Configure USB ethernet interface "wlan0" to run a DHCP server
#   - use IPv4 172.24.0.1 for interface with netmask 255.255.255.0
#   - disable DHCP option 3 (router) by passing an empty value
#   - disable DHCP option 6 (DNS) by passing an empty value
#   - add a DHCP range from 172.24.0.10 to 172.24.0.20 with a lease time of 5 minutes
P4wnP1_cli NET set server -i wlan0 -a 172.24.0.1 -m 255.255.255.0 -o "3:" -o "6:" -r "172.24.0.10|172.24.0.20|5m"

P4wnP1_cli LED set -b 5
