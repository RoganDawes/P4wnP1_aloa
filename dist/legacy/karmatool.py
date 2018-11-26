#!/usr/bin/python

#!/usr/bin/python

#    This file is part of P4wnP1.
#
#    Copyright (c) 2017, Marcus Mengs. 
#
#    P4wnP1 is free software: you can redistribute it and/or modify
#    it under the terms of the GNU General Public License as published by
#    the Free Software Foundation, either version 3 of the License, or
#    (at your option) any later version.
#
#    P4wnP1 is distributed in the hope that it will be useful,
#    but WITHOUT ANY WARRANTY; without even the implied warranty of
#    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#    GNU General Public License for more details.
#
#    You should have received a copy of the GNU General Public License
#    along with P4wnP1.  If not, see <http://www.gnu.org/licenses/>.


# The command line tool could be used to configure the MaMe82 nexmon firmware mod (KARMA)
# for Pi3 / Pi0W while an access point is up and running

from mame82_util import *
import cmd
import sys
import getopt

def interact():
	pass

def usage():
        usagescr = '''Firmware configuration tool for KARMA modified nexmon WiFi firmware on Pi0W/Pi3 by MaMe82
=========================================================================================

RePo:       https://github.com/mame82/P4wnP1_nexmon_additions
Creds to:   seemoo-lab for "NEXMON" project

A hostapd based Access Point should be up and running, when using this tool
(see the README for details).
            
Usage:      python karmatool.py [Arguments]

Arguments:
   -h                   Print this help screen
   -i                   Interactive mode
   -d                   Load default configuration (KARMA on, KARMA beaconing off, 
                        beaconing for 13 common SSIDs on, custom SSIDs never expire)
   -c                   Print current KARMA firmware configuration
   -p 0/1               Disable/Enable KARMA probe responses
   -a 0/1               Disable/Enable KARMA association responses
   -k 0/1               Disable/Enable KARMA association responses and probe responses
                        (overrides -p and -a)
   -b 0/1               Disable/Enable KARMA beaconing (broadcasts up to 20 SSIDs
                        spotted in probe requests as beacon)
   -s 0/1               Disable/Enable custom SSID beaconing (broadcasts up to 20 SSIDs
                        which have been added by the user with '--addssid=' when enabled)
   --addssid="test"     Add SSID "test" to custom SSID list (max 20 SSIDs)
   --remssid="test"     Remove SSID "test" from custom SSID list
   --clearssids         Clear list of custom SSIDs
   --clearkarma         Clear list of karma SSIDs (only influences beaconing, not probes)
   --autoremkarma=600   Auto remove KARMA SSIDs from beaconing list after sending 600 beacons
                        without receiving an association (about 60 seconds, 0 = beacon forever)
   --autoremcustom=3000    Auto remove custom SSIDs from beaconing list after sending 3000
                        beacons without receiving an association (about 5 minutes, 0 = beacon
                        forever)
   
Example:
   python karmatool.py -k 1 -b 0    Enables KARMA (probe and association responses)
                                    But sends no beacons for SSIDs from received probes
   python karmatool.py -k 1 -b 0    Enables KARMA (probe and association responses)
                                    and sends beacons for SSIDs from received probes
                                    (max 20 SSIDs, if autoremove isn't enabled)
   
   python karmatool.py --addssid="test 1" --addssid="test 2" -s 1
                                    Add SSID "test 1" and "test 2" and enable beaconing for
                                    custom SSIDs
'''
        print(usagescr)

def print_conf():
	print "Retrieving current configuration ...\n===================================="
	MaMe82_IO.dump_conf(print_res=True)
	
def check_bool_arg(arg):
	try:
		res = int(arg)
		if (res == 0) or (res == 1):
			return res
		else:
			return -1
	except ValueError:
		return -1

def main(argv):
	try:
		opts, args = getopt.getopt(argv, "hicdk:p:a:b:s:", ["help", "interactive", "currentconfig", "setdefault", "clearkarma", "clearssids", "addssid=", "remssid=", "autoremkarma=", "autoremcustom="])
	except getopt.GetoptError:
		print "ERROR: Wrong command line argument(s)"
		print "-------------------------------------\n"
		usage()
		sys.exit(2)
		
	for opt, arg in opts:
		if opt in ("-h", "--help"):
			usage()
			sys.exit()
		elif opt in ("-d", "--setdefault"):
			print "Setting default configuration ..."
			MaMe82_IO.set_defaults()
			print_conf()
			sys.exit()
		elif opt in ("-i", "--interactive"):
			print "Interactive mode"
			print "... Sorry, feature not implemented, yet ... stay tuned"
			sys.exit()
		elif opt in ("-c", "--currentconfig"):
			print_conf()
		elif opt == "-p":
			val = check_bool_arg(arg)
			if (val == -1):
				print "Argument error for -p (KARMA probe), must be 0 or 1 .... ignoring option"
			else:
				print "Setting KARMA probe responses to {0}".format("On" if (val==1) else "Off")
				MaMe82_IO.set_enable_karma_probe(True if (val==1) else False)
		elif opt == "-a":
			val = check_bool_arg(arg)
			if (val == -1):
				print "Argument error for -a (KARMA associations), must be 0 or 1 .... ignoring option"
			else:
				print "Setting KARMA association responses to {0}".format("On" if (val==1) else "Off")
				MaMe82_IO.set_enable_karma_assoc(True if (val==1) else False)
		elif opt == "-k":
			val = check_bool_arg(arg)
			if (val == -1):
				print "Argument error for -k (KARMA probes and associations), must be 0 or 1 .... ignoring option"
			else:
				print "Setting KARMA probe and association responses to {0}".format("On" if (val==1) else "Off")
				MaMe82_IO.set_enable_karma(True if (val==1) else False)
		elif opt == "-b":
			val = check_bool_arg(arg)
			if (val == -1):
				print "Argument error for -b (KARMA beaconing), must be 0 or 1 .... ignoring option"
			else:
				print "Setting KARMA beaconing to {0}".format("On" if (val==1) else "Off")
				MaMe82_IO.set_enable_karma_beaconing(True if (val==1) else False)
		elif opt == "-s":
			val = check_bool_arg(arg)
			if (val == -1):
				print "Argument error for -s (custom beaconing), must be 0 or 1 .... ignoring option"
			else:
				print "Setting custom beaconing to {0}".format("On" if (val==1) else "Off")
				MaMe82_IO.set_enable_custom_beaconing(True if (val==1) else False)
		elif opt == "--addssid":
			if len(arg) == 0 or len(arg) > 32:
				print "Argument error for --addssid, mustn't be empty max length is 32 ... ignoring option"
			else:
				MaMe82_IO.add_custom_ssid(arg)
		elif opt == "--remssid":
			if len(arg) == 0 or len(arg) > 32:
				print "Argument error for --remssid, mustn't be empty max length is 32 ... ignoring option"
			else:
				MaMe82_IO.rem_custom_ssid(arg)
		elif opt == "--clearssids":
			print "Removing all custom SSIDs"
			MaMe82_IO.clear_custom_ssids()
		elif opt == "--clearkarma":
			print "Removing all KARMA SSIDs (no influence on probe / assoc responses)"
			MaMe82_IO.clear_karma_ssids()
		elif opt == "--autoremkarma":
			error="An integer value >=0 is needed for autoremkarma ... ignoring option"
			try:
				val = int(arg)
				if (val < 0):
					print error
				else:
					print "Removing KARMA SSIDs after sending {0} beacons without occuring association".format(val)
					MaMe82_IO.set_autoremove_karma_ssids(val)
			except ValueError:
				print error
		elif opt == "--autoremcustom":
			error="An integer value >=0 is needed for autoremcustom ... ignoring option"
			try:
				val = int(arg)
				if (val < 0):
					print error
				else:
					print "Removing custom SSIDs after sending {0} beacons without occuring association".format(val)
					MaMe82_IO.set_autoremove_custom_ssids(val)
			except ValueError:
				print error
			
			
		
	print ""
	print_conf()

		

if __name__ == "__main__":
	if not MaMe82_IO.check_for_karma_cap():
		print "The current WiFi Firmware in use doesn't seem to support KARMA"
		print "A modified and precompiled nexmon firmware for Pi3 / Pi0w with KARMA support could"
		print "be found here:\thttps://github.com/mame82/P4wnP1_nexmon_additions"
		sys.exit()
	else:
		print "Firmware in use seems to be KARMA capable"
	
	if len(sys.argv) < 2:
		usage()
		sys.exit()
	main(sys.argv[1:])

