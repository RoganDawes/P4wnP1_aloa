/*
All the ugly stuff which not only depends on Linux (nothing is platform independent here), but
uses external binaries and depends on them (dnsmasq, dhclient, wpa_supplicant, hostapd etc.) ... or even
worse, the external binaries are glued together with /bin/bash tricks.
 */

package service
