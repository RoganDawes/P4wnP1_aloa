#!/bin/bash

# install dependencies
# - dnsmasq for DHCP / DNS server
# - bridge-utils for bonding CDC ECM + RNDIS interface toa single bridge
# - hostapd for AP deployment
# - screen to attach interactive processes to a detachable tty
# - autossh for "reachback" SSH connections
# - bluez (bluez-bleutooth, policykit-1) for access to Bluetooth / BLE stack (depends on DBUS systemd service)
# - haveged as entropy daemon to get enough entropy for hostapd AP with WPA2
# - iodine for DNS tunnel capbilities
# - genisoimage to allow on-the-fly CD-Rom image creation for CD emulation

sudo apt-get -y install git screen hostapd autossh bluez bluez-tools bridge-utils policykit-1 genisoimage iodine haveged
sudo apt-get -y install tcpdump
sudo apt-get -y install python-pip python-dev

# before installing dnsmasq, the nameserver from /etc/resolv.conf should be saved
# to restore after install (gets overwritten by dnsmasq package)
cp /etc/resolv.conf /tmp/backup_resolv.conf
sudo apt-get -y install dnsmasq
sudo /bin/bash -c 'cat /tmp/backup_resolv.conf > /etc/resolv.conf'



# python dependencies for HIDbackdoor
sudo pip install pycrypto # already present on stretch
sudo pip install pydispatcher


