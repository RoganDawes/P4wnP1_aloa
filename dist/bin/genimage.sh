#!/bin/bash

# requires genisoimage package
# usage: genimage.sh <outfile> <type = cdrom || flashdrive> [volume label] [size] 


ISO_FOLDER="/tmp/iso"
VOL_ID="Test_CD"
OUTFILE="/tmp/cdrom.iso"
OUTFILE2="/tmp/image.bin"
size="128" # only used for flashdrive, given in Megabyte

function create_cdrom_image() {
	rm -R $ISO_FOLDER # in case it exists
	mkdir $ISO_FOLDER
	printf "Hello World!\r\nP4wnP1" > $ISO_FOLDER/hello.txt

	# generate iso
	genisoimage -udf -joliet-long -V $VOL_ID -o $OUTFILE $ISO_FOLDER
}

function create_block_vfat_image() {
	dd if=/dev/zero of=$OUTFILE2 bs=1M count=$size
	#mkdosfs $OUTFILE # create vfat
	mk.vfat $OUTFILE # create vfat
}

function loop_mount() {
	# find free loop device
	loopdev=$(losetup -f)
	
	losetup $loopdev $OUTFILE2
	
	# mounting the image to /mnt
	# Note: If the image is used by USB Mass Storage currently, the behavior is unpredictable
	#       Based on observation, local changes to the mounted block device have no effect, but 
	#       (external) changes to the USB Mass Storage have an effect. This behavior doesn't change
	#       if the loop mount is done with Direct-IO or the loop device is mounted with -o=sync.
	#
	#		In addtion, binding vfat images to USB Mass Storage with CD-Rom emulation doesn't work,
	#		it seems the filesystem has to be ISO9660.
	#
	#		Last but not least, even if the USB Mass Storage is flagged as removable, resetting the
	#		backing file to "" only works if the target host hasn't mounted the image (very unlikely
	#		with automount, like on Windows). There's no "forced" unmount possible from P4wnP1's end,
	#		therefor the whole gadget hast to be disabled and re-enabled (not reinitialized) to bring 
	#		up another volume. This will interrupt other USB functions enabled on the current composite gadget.
	#
	# Conclusion:
	#		The backing service should allow mounting of image files only to loopback OR USB Mass Storage,
	#		not both.
	#		UMS backing image file and operation mode will be integrated into gadgetsettings (could only
	#		be changed on reinitialization) because of the observation according removable devices and the
	#		lacking possibility to change the backing file, once the device is mounted by the target host.
	mount -t vfat -o loop $OUTFILE /mnt
	
	# to detach loopdefv:
	#	losetup -d $loopdev
	
	# to mount image filesystem
	#	mount -t vfat -o loop $OUTFILE2 $MOUNTPOINT
}

create_cdrom_image
create_block_vfat_image
