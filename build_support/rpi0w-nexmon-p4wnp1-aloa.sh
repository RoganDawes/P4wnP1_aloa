#!/bin/bash
set -e

######
# This is a work in progress script for P4wnP1 A.L.O.A. based on @binkybear's built script for P4wnP1
#
##########


if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root"
   exit 1
fi

if [[ $# -eq 0 ]] ; then
    echo "Please pass version number, e.g. $0 2.0"
    exit 0
fi

basedir=`pwd`/rpi0w-nexmon-p4wnp1-aloa-$1
TOPDIR=`pwd`

# Custom hostname variable
hostname=${2:-kali}
# Custom image name variable - MUST NOT include .img at the end.
imagename=${3:-kali-linux-$1-rpi0w-nexmon-p4wnp1-aloa}
# Size of image in megabytes (Default is 4500=4.5GB)
size=6000
# Suite to use.  
# Valid options are:
# kali-rolling, kali-dev, kali-bleeding-edge, kali-dev-only, kali-experimental, kali-last-snapshot
# A release is done against kali-last-snapshot, but if you're building your own, you'll probably want to build
# kali-rolling.
suite=kali-rolling

# Generate a random machine name to be used.
machine=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 16 | head -n 1)

# Package installations for various sections.
# This will build a minimal Kali system with the top 10 tools.
# This is the section to edit if you would like to add more packages.
# See http://www.kali.org/new/kali-linux-metapackages/ for meta packages you can
# use. You can also install packages, using just the package name, but keep in
# mind that not all packages work on ARM! If you specify one of those, the
# script will throw an error, but will still continue on, and create an unusable
# image, keep that in mind.

arm="fake-hwclock ntpdate u-boot-tools"
tools="aircrack-ng crunch cewl dnsrecon dnsutils ethtool exploitdb hydra medusa metasploit-framework ncrack nmap passing-the-hash proxychains recon-ng sqlmap tcpdump theharvester tor tshark usbutils whois windows-binaries winexe wpscan"
base="apt-transport-https apt-utils console-setup e2fsprogs firmware-linux firmware-realtek firmware-atheros ifupdown initramfs-tools iw kali-defaults man-db mlocate netcat-traditional net-tools parted psmisc rfkill screen snmpd snmp tftp tmux unrar usbutils vim wget zerofree"
services="apache2 atftpd openssh-server openvpn"
# haveged: assure enough entropy data for hostapd on startup
# avahi-daemon: allow mDNS resolution (apple bonjour) by remote hosts
# dhcpcd5: REQUIRED (P4wnP1 A.L.O.A. currently wraps this binary if a DHCP client is needed)
# dnsmasq: REQUIRED (P4wnP1 A.L.O.A. currently wraps this binary if a DHCP server is needed, currently not used for DNS)
# genisoimage: allow creation of CD-Rom iso images for CD-Rom USB gadget from existing folders on the fly
# iodine: allow DNS tunneling
# dosfstools: contains fatlabel (used to label FAT32 iamges for UMS)
# Note on Go: The golang package is version 1.10, so we are missing support for current gopherjs (webclient couldn't be build on Pi) and go modules (replacement for dep)
extras="autossh avahi-daemon bash-completion dhcpcd5 dnsmasq dosfstools genisoimage golang haveged hostapd i2c-tools iodine policykit-1 python-configobj python-dev python-pip python-requests wpasupplicant"

packages="${arm} ${base} ${services} ${extras}"
architecture="armel"
# If you have your own preferred mirrors, set them here.
# After generating the rootfs, we set the sources.list to the default settings.
mirror=http.kali.org

# Check to ensure that the architecture is set to ARMEL since the RPi is the
# only board that is armel.
if [[ ${architecture} != "armel" ]] ; then
    echo "The Raspberry Pi cannot run Debian armhf binaries"
    exit 0
fi

# Set this to use an http proxy, like apt-cacher-ng, and uncomment further down
# to unset it.
#export http_proxy="http://localhost:3142/"

mkdir -p "${basedir}"
cd "${basedir}"

# create the rootfs - not much to modify here, except maybe throw in some more packages if you want.
debootstrap --foreign --keyring=/usr/share/keyrings/kali-archive-keyring.gpg --include=kali-archive-keyring --arch ${architecture} ${suite} kali-${architecture} http://${mirror}/kali

LANG=C systemd-nspawn -M ${machine} -D kali-${architecture} /debootstrap/debootstrap --second-stage

mkdir -p kali-${architecture}/etc/apt/
cat << EOF > kali-${architecture}/etc/apt/sources.list
deb http://${mirror}/kali ${suite} main contrib non-free
EOF

# Set hostname
echo "${machine}" > kali-${architecture}/etc/hostname

# So X doesn't complain, we add $hostname to hosts
cat << EOF > kali-${architecture}/etc/hosts
127.0.0.1       ${machine}    localhost
::1             localhost ip6-localhost ip6-loopback
fe00::0         ip6-localnet
ff00::0         ip6-mcastprefix
ff02::1         ip6-allnodes
ff02::2         ip6-allrouters
EOF

# added wlan0 configuration to allow copying of wpa_supplicant.conf to work
mkdir -p kali-${architecture}/etc/network/
cat << EOF > kali-${architecture}/etc/network/interfaces
auto lo
iface lo inet loopback

EOF

cat << EOF > kali-${architecture}/etc/resolv.conf
nameserver 8.8.8.8
EOF

export MALLOC_CHECK_=0 # workaround for LP: #520465
export LC_ALL=C
export DEBIAN_FRONTEND=noninteractive

#mount -t proc proc kali-$architecture/proc
#mount -o bind /dev/ kali-$architecture/dev/
#mount -o bind /dev/pts kali-$architecture/dev/pts

cat << EOF > kali-${architecture}/debconf.set
console-common console-data/keymap/policy select Select keymap from full list
console-common console-data/keymap/full select en-latin1-nodeadkeys
EOF

# Create monitor mode start/remove
# The script returns an error code if the monitor interface couldn't be started
# Note: Removing this should be considered, as enabling the monitor interface once
# and using wpa_supplicant afterwards, crashs the WiFi firmware (even if the monitor
# interface is removed). Afterwards the 'brcmfmac' module has to be removed and
# loaded again (the driver push the firmware and restarts the fmac chip on init).
# Sometimes only a reboot works
cat << 'EOF' > kali-${architecture}/usr/bin/monstart
#!/bin/bash
interface=wlan0mon
echo -n "Create monitor mode interface ${interface}... "
iw phy phy0 interface add ${interface} type monitor 2> /dev/null 1> /dev/null
if [ $? -eq 0 ]; then
  echo "success"
else
  echo "failed, already created ?"
fi

echo -n "Trying to enable ${interface}... "
ifconfig ${interface} up 2> /dev/null
if [ $? -eq 0 ]; then
  echo "success, ${interface} is up"
  exit 0
else
  echo "failed"
  exit 1
fi
EOF
chmod 755 kali-${architecture}/usr/bin/monstart

cat << EOF > kali-${architecture}/usr/bin/monstop
#!/bin/bash
interface=wlan0mon
ifconfig ${interface} down
sleep 1
iw dev ${interface} del
EOF
chmod 755 kali-${architecture}/usr/bin/monstop

mkdir -p kali-${architecture}/lib/systemd/system/
cat << 'EOF' > kali-${architecture}/lib/systemd/system/regenerate_ssh_host_keys.service
[Unit]
Description=Regenerate SSH host keys
Before=ssh.service
[Service]
Type=oneshot
ExecStartPre=-/bin/dd if=/dev/hwrng of=/dev/urandom count=1 bs=4096
ExecStartPre=-/bin/sh -c "/bin/rm -f -v /etc/ssh/ssh_host_*_key*"
ExecStart=/usr/bin/ssh-keygen -A -v
ExecStartPost=/bin/sh -c "for i in /etc/ssh/ssh_host_*_key*; do actualsize=$(wc -c <\"$i\") ;if [ $actualsize -eq 0 ]; then echo size is 0 bytes ; exit 1 ; fi ; done ; /bin/systemctl disable regenerate_ssh_host_keys"
[Install]
WantedBy=multi-user.target
EOF
chmod 644 kali-${architecture}/lib/systemd/system/regenerate_ssh_host_keys.service

cat << EOF > kali-${architecture}/lib/systemd/system/rpiwiggle.service
[Unit]
Description=Resize filesystem
Before=regenerate_ssh_host_keys.service
[Service]
Type=oneshot
ExecStart=/root/scripts/rpi-wiggle.sh
ExecStartPost=/bin/systemctl disable rpiwiggle
[Install]
WantedBy=multi-user.target
EOF
chmod 644 kali-${architecture}/lib/systemd/system/rpiwiggle.service

cat << EOF > "${basedir}"/kali-${architecture}/lib/systemd/system/enable-ssh.service
[Unit]
Description=Turn on SSH if /boot/ssh is present
ConditionPathExistsGlob=/boot/ssh{,.txt}
After=regenerate_ssh_host_keys.service

[Service]
Type=oneshot
ExecStart=/bin/sh -c "update-rc.d ssh enable && invoke-rc.d ssh start && rm -f /boot/ssh ; rm -f /boot/ssh.txt"

[Install]
WantedBy=multi-user.target
EOF
chmod 644 "${basedir}"/kali-${architecture}/lib/systemd/system/enable-ssh.service

# Bluetooth enabling
mkdir -p kali-${architecture}/lib/udev/rules.d/
cp "${basedir}"/../misc/pi-bluetooth/50-bluetooth-hci-auto-poweron.rules kali-${architecture}/lib/udev/rules.d/50-bluetooth-hci-auto-poweron.rules
cp "${basedir}"/../misc/pi-bluetooth/pi-bluetooth+re4son_2.2_all.deb kali-${architecture}/root/pi-bluetooth+re4son_2.2_all.deb

# Copy a default config, with everything commented out so people find it when
# they go to add something when they are following instructions on a website.
cp "${basedir}"/../misc/config.txt "${basedir}"/kali-${architecture}/boot/config.txt

# move P4wnP1 in (change to release blob when ready)
git clone  -b 'v0.1.0-alpha2' --single-branch --depth 1  https://github.com/mame82/P4wnP1_aloa "${basedir}"/kali-${architecture}/root/P4wnP1

cat << EOF > kali-${architecture}/third-stage
#!/bin/bash
set -e
dpkg-divert --add --local --divert /usr/sbin/invoke-rc.d.chroot --rename /usr/sbin/invoke-rc.d
cp /bin/true /usr/sbin/invoke-rc.d
echo -e "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d
chmod 755 /usr/sbin/policy-rc.d

apt-get update
apt-get --yes --allow-change-held-packages install locales-all

debconf-set-selections /debconf.set
rm -f /debconf.set
apt-get update
apt-get -y install git-core binutils ca-certificates initramfs-tools u-boot-tools ssh
apt-get -y install locales console-common less nano git
echo "root:toor" | chpasswd
rm -f /etc/udev/rules.d/70-persistent-net.rules
export DEBIAN_FRONTEND=noninteractive
apt-get --yes --allow-change-held-packages install ${packages} || apt-get --yes --fix-broken install
apt-get --yes --allow-change-held-packages install bluez-firmware bluez libasound2 libasound2-data libdw1 || apt-get --yes --fix-broken install
apt-get --yes --allow-change-held-packages install ${desktop} ${tools} || apt-get --yes --fix-broken install
apt-get --yes --allow-change-held-packages dist-upgrade
apt-get --yes --allow-change-held-packages autoremove

# Because copying in authorized_keys is hard for people to do, let's make the
# image insecure and enable root login with a password.
echo "Allow root login..."
sed -i -e 's/^#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

# Resize FS on first run (hopefully)
systemctl enable rpiwiggle

# Generate SSH host keys on first run
systemctl enable regenerate_ssh_host_keys
systemctl enable ssh

# Install and hold pi-bluetooth deb package from re4son
dpkg --force-all -i /root/pi-bluetooth+re4son_2.2_all.deb
apt-mark hold pi-bluetooth+re4son

# systemd version 232 and above breaks execution of above bluetooth rule, let's fix that
sed -i 's/^RestrictAddressFamilies=AF_UNIX AF_NETLINK AF_INET AF_INET6.*/RestrictAddressFamilies=AF_UNIX AF_NETLINK AF_INET AF_INET6 AF_BLUETOOTH/' /lib/systemd/system/systemd-udevd.service

# Enable bluetooth
systemctl unmask bluetooth.service
systemctl enable bluetooth
systemctl enable hciuart
# dhcpcd is needed by P4wnP1, but started on demand
# installation of dhcpcd5 package enables a systemd unit starting dhcpcd for all
# interfaces, which results in conflicts with DHCP servers running on created
# bridge interface (especially for the bteth BNEP bridge). To avoid this we
# disable the service. If communication problems occur, although DHCP leases
# are handed out by dnsmasq, dhcpcd should be the first place to look
# (no interface should hava an APIPA addr assigned, unless the DHCP client
# was explcitely enabled by P4wnP1 for this interface)
systemctl disable dhcpcd

# enable fake-hwclock (P4wnP1 is intended to reboot/loose power frequently without getting NTP access in between)
# a clean shutdown/reboot is needed, as fake-hwclock service saves time on stop
systemctl enable fake-hwclock

# Create cmdline.txt file
mkdir -p /boot
echo "dwc_otg.lpm_enable=0 console=serial0,115200 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 elevator=deadline fsck.repair=yes rootwait" > /boot/cmdline.txt

# Install P4wnP1 A.L.O.A.
cd /root/P4wnP1
make installkali

# add Designware DUAL role USB driver to loaded modules
echo "dwc2" | tee -a /etc/modules

# allow root login from tyyGS0 (serial device for USB gadget)
echo ttyGS0 >> /etc/securetty

# add minutely cronjob to update fake-hwclock
echo '* * * * * root /usr/sbin/fake-hwclock' >> /etc/crontab

# Turn off kernel dmesg showing up in console since rpi0 only uses console
echo "dmesg -D" > /etc/rc.local
echo "exit 0" >> /etc/rc.local

# Copy bashrc
cp  /etc/skel/.bashrc /root/.bashrc

# libinput seems to fail hard on RaspberryPi devices, so we make sure it's not
# installed here (and we have xserver-xorg-input-evdev and
# xserver-xorg-input-synaptics packages installed above!)
apt-get --yes --allow-change-held-packages purge xserver-xorg-input-libinput

# Fix startup time from 5 minutes to 15 secs on raise interface wlan0
sed -i 's/^TimeoutStartSec=5min/TimeoutStartSec=15/g' "/lib/systemd/system/networking.service"

rm -f /usr/sbin/policy-rc.d
rm -f /usr/sbin/invoke-rc.d
dpkg-divert --remove --rename /usr/sbin/invoke-rc.d

rm -f /third-stage
EOF

chmod 755 kali-${architecture}/third-stage
LANG=C systemd-nspawn -M ${machine} -D kali-${architecture} /third-stage
if [[ $? > 0 ]]; then
  echo "Third stage failed"
  exit 1
fi

cat << EOF > kali-${architecture}/cleanup
#!/bin/bash
rm -rf /root/.bash_history
rm -rf /root/P4wnP1_go
apt-get update
apt-get clean
rm -f /0
rm -f /hs_err*
rm -f cleanup
rm -f /usr/bin/qemu*
EOF

chmod 755 kali-${architecture}/cleanup
LANG=C systemd-nspawn -M ${machine} -D kali-${architecture} /cleanup

# Enable login over serial
echo "T0:23:respawn:/sbin/agetty -L ttyAMA0 115200 vt100" >> "${basedir}"/kali-${architecture}/etc/inittab

cat << EOF > "${basedir}"/kali-${architecture}/etc/apt/sources.list
deb http://http.kali.org/kali kali-rolling main non-free contrib
deb-src http://http.kali.org/kali kali-rolling main non-free contrib
EOF

# Uncomment this if you use apt-cacher-ng otherwise git clones will fail.
#unset http_proxy

# Kernel section. If you want to use a custom kernel, or configuration, replace
# them in this section.

cd ${TOPDIR}

# RPI Firmware
git clone --depth 1 https://github.com/raspberrypi/firmware.git rpi-firmware
cp -rf rpi-firmware/boot/* "${basedir}"/kali-${architecture}/boot/
# copy over Pi specific libs (video core) and binaries (dtoverlay,dtparam ...)
cp -rf rpi-firmware/opt/* "${basedir}"/kali-${architecture}/opt/
rm -rf rpi-firmware

# Build nexmon firmware outside the build system, if we can (use repository with driver and firmware for P4wnP1).
cd "${basedir}"
git clone https://github.com/mame82/nexmon_wifi_covert_channel.git -b p4wnp1 "${basedir}"/nexmon --depth 1

# Setup build
cd ${TOPDIR}
# Re4son kernel 4.14.80 with P4wnP1 patches (dwc2 and brcmfmac)
git clone --depth 1 https://github.com/Re4son/re4son-raspberrypi-linux -b rpi-4.14.80-re4son-p4wnp1 "${basedir}"/kali-${architecture}/usr/src/kernel


cd "${basedir}"/kali-${architecture}/usr/src/kernel

# Note: Compiling the kernel in /usr/src/kernel of the target file system is problematic, as the binaries of the compiling host architecture
# get deployed to the /usr/src/kernel/scripts subfolder (in this case linux-x64 binaries), which is symlinked to /usr/src/build later on.
# This would f.e. hinder rebuilding single modules, like nexmon's brcmfmac driver, on the Pi itself (online compilation). 
# The cause:building of modules relies on the pre-built binaries in /usr/src/build folder. But the helper binaries are compiled with the
# HOST toolchain and not with the crosscompiler toolchain (f.e. /usr/src/kernel/script/basic/fixdep would end up as x64 binary, as this helper
# is not compiled with the CROSS toolchain). As those scripts are used druing module build, it wouldn't work to build on the pi, later on,
# without recompiling the helper binaries with the proper crosscompiler toolchain.
#
# To account for that, the 'script' subfolder could be rebuild on the target (online) by running `make scripts/` from /usr/src/kernel folder.
# Rebuilding the script, again, depends on additional tooling, like `bc` binary, which has to be installed.
#
# Currently the step of recompiling the kernel/scripts folder has to be done manually online, but it should be possible to do it after kernel
# build, by setting the host compiler (CC) to the gcc of the linaro-arm-linux-gnueabihf-raspbian-x64 toolchain (not only the CROSS_COMPILE).
# The problem is, that the used linaro toolchain builds for armhf (not a problem for kernel, as there're no dependencies on hf librearies),
# but the debian packages (and the provided gcc) are armel.
#
# To clean up this whole "armel" vs "armhf" mess, the kernel should be compiled with a armel toolchain (best choice would be the toolchain
# which is used to build the kali armel packages itself, which is hopefully available for linux-x64)
#
# For now this is left as manual step, as the normal user shouldn't have a need to recompile kernel parts on the Pi itself.


# Set default defconfig
export ARCH=arm
# use hard float with RPi cross compiler toolchain, as described here: https://www.raspberrypi.org/documentation/linux/kernel/building.md
export CROSS_COMPILE=arm-linux-gnueabi-

# Set default defconfig
make re4son_pi1_defconfig

# Build kernel
make -j $(grep -c processor /proc/cpuinfo)

# Make kernel modules
make modules_install INSTALL_MOD_PATH="${basedir}"/kali-${architecture}

# Copy kernel to boot
perl scripts/mkknlimg --dtok arch/arm/boot/zImage "${basedir}"/kali-${architecture}/boot/kernel.img
cp arch/arm/boot/dts/*.dtb "${basedir}"/kali-${architecture}/boot/
cp arch/arm/boot/dts/overlays/*.dtb* "${basedir}"/kali-${architecture}/boot/overlays/
cp arch/arm/boot/dts/overlays/README "${basedir}"/kali-${architecture}/boot/overlays/

make mrproper
make re4son_pi1_defconfig

# Fix up the symlink for building external modules
# kernver is used so we don't need to keep track of what the current compiled
# version is
kernver=$(ls "${basedir}"/kali-${architecture}/lib/modules/)
cd "${basedir}"/kali-${architecture}/lib/modules/${kernver}
rm build
rm source
ln -s /usr/src/kernel build
ln -s /usr/src/kernel source
cd "${basedir}"

# Copy a default config, with everything commented out so people find it when
# they go to add something when they are following instructions on a website.
cp "${basedir}"/../misc/config.txt "${basedir}"/kali-${architecture}/boot/config.txt

cat << EOF >> "${basedir}"/kali-${architecture}/boot/config.txt
dtoverlay=dwc2
EOF

# systemd doesn't seem to be generating the fstab properly for some people, so
# let's create one.
cat << EOF > "${basedir}"/kali-${architecture}/etc/fstab
# <file system> <mount point>   <type>  <options>       <dump>  <pass>
proc            /proc           proc    defaults          0       0
/dev/mmcblk0p1  /boot           vfat    defaults          0       2
/dev/mmcblk0p2  /               ext4    defaults,noatime  0       1
EOF


# rpi-wiggle
mkdir -p "${basedir}"/kali-${architecture}/root/scripts
wget https://raw.github.com/steev/rpiwiggle/master/rpi-wiggle -O "${basedir}"/kali-${architecture}/root/scripts/rpi-wiggle.sh
chmod 755 "${basedir}"/kali-${architecture}/root/scripts/rpi-wiggle.sh

# git clone of nexmon moved in front of kernel compilation, to have poper brcmfmac driver ready
cd "${basedir}"/nexmon
# Make sure we're not still using the armel cross compiler
unset CROSS_COMPILE

# Disable statistics
touch DISABLE_STATISTICS
source setup_env.sh
make
cd buildtools/isl-0.10
CC=$CCgcc
./configure
make
sed -i -e 's/all:.*/all: $(RAM_FILE)/g' ${NEXMON_ROOT}/patches/bcm43430a1/7_45_41_46/nexmon/Makefile
cd ${NEXMON_ROOT}/patches/bcm43430a1/7_45_41_46/nexmon
make clean
# We do this so we don't have to install the ancient isl version into /usr/local/lib on systems.
LD_LIBRARY_PATH=${NEXMON_ROOT}/buildtools/isl-0.10/.libs make ARCH=arm CC=${NEXMON_ROOT}/buildtools/gcc-arm-none-eabi-5_4-2016q2-linux-x86/bin/arm-none-eabi-
# RPi0w->3B firmware
# disable nexmon by default
mkdir -p "${basedir}"/kali-${architecture}/lib/firmware/brcm
cp ${NEXMON_ROOT}/patches/bcm43430a1/7_45_41_46/nexmon/brcmfmac43430-sdio.bin "${basedir}"/kali-${architecture}/lib/firmware/brcm/brcmfmac43430-sdio.nexmon.bin
cp ${NEXMON_ROOT}/patches/bcm43430a1/7_45_41_46/nexmon/brcmfmac43430-sdio.bin "${basedir}"/kali-${architecture}/lib/firmware/brcm/brcmfmac43430-sdio.bin
wget https://raw.githubusercontent.com/RPi-Distro/firmware-nonfree/master/brcm/brcmfmac43430-sdio.txt -O "${basedir}"/kali-${architecture}/lib/firmware/brcm/brcmfmac43430-sdio.txt
# Make a backup copy of the rpi firmware in case people don't want to use the nexmon firmware.
# The firmware used on the RPi is not the same firmware that is in the firmware-brcm package which is why we do this.
wget https://raw.githubusercontent.com/RPi-Distro/firmware-nonfree/master/brcm/brcmfmac43430-sdio.bin -O "${basedir}"/kali-${architecture}/lib/firmware/brcm/brcmfmac43430-sdio.rpi.bin
#cp "${basedir}"/kali-${architecture}/lib/firmware/brcm/brcmfmac43430-sdio.rpi.bin "${basedir}"/kali-${architecture}/lib/firmware/brcm/brcmfmac43430-sdio.bin

cp "${basedir}"/../misc/brcm/BCM43430A1.hcd "${basedir}"/kali-${architecture}/lib/firmware/brcm/BCM43430A1.hcd

cd "${basedir}"

cp "${basedir}"/../misc/zram "${basedir}"/kali-${architecture}/etc/init.d/zram
chmod 755 "${basedir}"/kali-${architecture}/etc/init.d/zram

sed -i -e 's/^#PermitRootLogin.*/PermitRootLogin yes/' "${basedir}"/kali-${architecture}/etc/ssh/sshd_config

# Create the disk and partition it
echo "Creating image file ${imagename}.img"
dd if=/dev/zero of="${basedir}"/${imagename}.img bs=1M count=${size}
parted ${imagename}.img --script -- mklabel msdos
parted ${imagename}.img --script -- mkpart primary fat32 0 64
parted ${imagename}.img --script -- mkpart primary ext4 64 -1

# Set the partition variables
loopdevice=`losetup -f --show "${basedir}"/${imagename}.img`
device=`kpartx -va ${loopdevice} | sed 's/.*\(loop[0-9]\+\)p.*/\1/g' | head -1`
sleep 5
device="/dev/mapper/${device}"
bootp=${device}p1
rootp=${device}p2

# Create file systems
mkfs.vfat ${bootp}
mkfs.ext4 ${rootp}

# Create the dirs for the partitions and mount them
mkdir -p "${basedir}"/root
mount ${rootp} "${basedir}"/root
mkdir -p "${basedir}"/root/boot
mount ${bootp} "${basedir}"/root/boot

# We do this down here to get rid of the build system's resolv.conf after running through the build.
cat << EOF > kali-${architecture}/etc/resolv.conf
nameserver 8.8.8.8
EOF

# Because of the p4wnp1 script, we set the hostname down here, instead of using the machine name.
# Set hostname
echo "${hostname}" > "${basedir}"/kali-${architecture}/etc/hostname

# So X doesn't complain, we add $hostname to hosts
cat << EOF > "${basedir}"/kali-${architecture}/etc/hosts
127.0.0.1       ${hostname}    localhost
::1             localhost ip6-localhost ip6-loopback
fe00::0         ip6-localnet
ff00::0         ip6-mcastprefix
ff02::1         ip6-allnodes
ff02::2         ip6-allrouters
EOF

echo "Rsyncing rootfs into image file"
rsync -HPavz -q "${basedir}"/kali-${architecture}/ "${basedir}"/root/

# Unmount partitions
sync
umount ${bootp}
umount ${rootp}
kpartx -dv ${loopdevice}
losetup -d ${loopdevice}

# Don't pixz on 32bit, there isn't enough memory to compress the images.
MACHINE_TYPE=`uname -m`
if [ ${MACHINE_TYPE} == 'x86_64' ]; then
echo "Compressing ${imagename}.img"
pixz "${basedir}"/${imagename}.img "${basedir}"/../${imagename}.img.xz
fi

# Clean up all the temporary build stuff and remove the directories.
# Comment this out to keep things around if you want to see what may have gone
# wrong.
echo "Cleaning up the temporary build files..."
rm -rf "${basedir}"
