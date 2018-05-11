# Update USB gadget configuration (enable RNDIS, disable CDC ECM, enable HID Keyboard)
P4wnP1_cli USB set -r1 -e0 -k1

# Set new network configuration for "usbeth" (the interface is present when USB RNDIS, CDC ECM or both are enabled)
# - "server" means a DHCP server is started for the interface "usbeth"
# - set the address of the interface to 172.16.0.1 (-a flag)
# - set the netmask of the interface to 255.255.255.252 (-m flag)
# - add a range 127.16.0.2 to 172.16.0.2 to the DHCP server with leastime 3 minutes (-r flag, could be used multiple times to add more ranges)
# - add option 3 (ROUTER) to the DHCP server, but don't provide a value to disable sending a gateway entry (-o flag)
# - add option 6 (NAMESERVER) to the DHCP server, but don't provide a value to disable sending a DNS entry (-o flag, again)
# - add option 252 (WPAD) to DHCP server with value 'http://172.16.0.1/wpad.dat' (-o flag, again)
P4wnP1_cli NET set server -i usbeth -a 172.16.0.1 -m 255.255.255.252 -r "172.16.0.2|172.16.0.2|3m" -o "3:" -o "6:" -o "252:http://172.16.0.1/wpad.dat"

# Note: valid DHCP options are defined in RFC 2132 and additional RFCs (f.e. draft-ietf-wrec-wpad-01 defines WPAD)
# Note 2: some option values are lists with comma, f.e option 121 (static route) "121:10.0.0.0/8,10.0.0.1,11.0.0.0,10.0.0.1"
#         as the comma "," is already used as delimiter for multiple options, it has to be replaced by a pipe operator "|"
#         and the option has to be provided like this:
#                  -o "121:10.0.0.0/8|10.0.0.1|11.0.0.0/8|10.0.0.1"



# Start a DHCP Client for interface wlan0
P4wnP1_cli NET set client -i wlan0
