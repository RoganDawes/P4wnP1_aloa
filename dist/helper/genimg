#!/bin/bash

usage() {
	echo "genimg"
	echo "-------------------"
	echo "Generates FAT32 or ISO CD-Rom images for P4wnP1 A.L.O.A."
	echo "USB mass storage emulation"
	echo
	echo "Usage: genimage -i <folder> -o <imagename>"
	echo
	echo "Options:"
	echo
	echo "  -h, --help"
	echo "      This help text."
	echo
	echo "  -c, --cdrom"
	echo "      Build UDF joilet ISO image, if not given build FAT32 image."
	echo
	echo "  -l <string>, --label <string>"
	echo "      Used as volume ID for ISO image or drive label for FAT32"
	echo
	echo "  -s <number>, --size <number>"
	echo "      Image size in MByte (applies only to FAT32 image)"
	echo
	echo "  -i <folder>, --input <folder>"
	echo "      Input folder used to build the CD-Rom image."
	echo "      Optional for FAT32 iamge, if given content is copied."
	echo
	echo "  -o <imagename>, --output <imagename>"
	echo "      Output file name (without extension and path)."
	echo
}

if [ "$#" -eq 0 ]; then
	usage
	exit
fi

# defaults
label="P4wnP1 ALOA"
size=128
cdrom=false
ISO_PATH="/usr/local/P4wnP1/ums/cdrom"
FAT_PATH="/usr/local/P4wnP1/ums/flashdrive"

while [ "$#" -gt 0 ]
do
	case "$1" in
	-h|--help)
		usage
		exit 0
		;;
	-c|--cdrom)
		cdrom=true
		;;
	-i|--input)
		input="$2"
		;;
	-l|--label)
		label="$2"
		;;
	-s|--size)
		size="$2"
		;;
	-o|--output)
		output="$2"
		;;
	-*)
		echo "Invalid option '$1'. Use --help to see the valid options" >&2
		exit 1
		;;
	# an option argument, continue
	*)	;;
	esac
	shift
done

if $cdrom; then
	OUTFILE="$ISO_PATH/$output.iso"
	echo "Generating ISO image from $input at $OUTFILE"
	genisoimage -udf -joliet-long -V "$label" -o $OUTFILE $input
else
	OUTFILE="$FAT_PATH/$output.bin"
	echo "Generating $size""MB FAT32 image at $OUTFILE"
	dd if=/dev/zero of=$OUTFILE bs=1M count=$size
	mkdosfs $OUTFILE
	fatlabel $OUTFILE "$label"

	if [ "$input" != "" ]; then
		echo "Copying in input from $input"
		# find free loop device
	        loopdev=$(losetup -f)
		# bind image to loop device
		losetup $loopdev $OUTFILE
		# mount to /mnt
		mount -t vfat -o loop $OUTFILE /mnt

		# copy files and subfolders
		cp -R $input/* /mnt

		umount /mnt
		losetup -d $loopdev
	fi
fi
