#! /usr/bin/env sh
# Copyright (c) cmj <cmj@cmj.tw>. All right reserved.

## Global Variable ##
HOSTNAME=PwnPi
IMG_NAME=ArchLinuxARM-rpi-latest.tar.gz
ARCH_LINUX_IMG=http://os.archlinuxarm.org/os/${IMG_NAME}
LOCAL_IMG=/tmp/${IMG_NAME}
BOOT=boot
ROOT=root


DEV=$1

function get_last_image() {
	if [ ! -f "${LOCAL_IMG}" ]; then
		wget "${ARCH_LINUX_IMG}" -O "${LOCAL_IMG}"
	fi

	# verify the MD5 checksum
	wget ${ARCH_LINUX_IMG}.md5 -qO - | sed "s%${IMG_NAME}%${LOCAL_IMG}%g" | md5sum -c
	if [ $? != 0 ]; then
		echo "MD5 checksum not correct ..."
		exit -1
	fi
}

function set_partition() {
	DEV="$1"

	if [ ! -b "${DEV}" ]; then
		echo "${DEV} is not the block device"
		exit -1
	fi

	parted -s ${DEV} mklabel msdos
	parted -s ${DEV} mkpart primary fat32 1 128
	parted -s ${DEV} mkpart primary ext4 --  128 -1
	mkfs.vfat ${DEV}1
	mkfs.ext4 -F ${DEV}2
}

function enable_systemd {
	mkdir -p $(dirname ${ROOT}/etc/systemd/$2)
	ln -s /usr/lib/systemd/$1 ${ROOT}/etc/systemd/$2
}

function teardown() {
    umount ${BOOT} || true
    umount ${ROOT} || true
}

if [ `id -u` != 0 ]; then
	echo "Need root permission!"
	exit -1
fi

get_last_image
set_partition "${DEV}"

# auto-umount
trap teardown EXIT

mkdir -p ${BOOT} ${ROOT}
mount ${DEV}1 ${BOOT}
mount ${DEV}2 ${ROOT}

bsdtar -xpf ${LOCAL_IMG} -C ${ROOT}
sync
mv ${ROOT}/${BOOT}/* ${BOOT}/

# install the extra configuration #
echo "${HOSTNAME}" > ${ROOT}/etc/hostname

## enable dwc2 kernel module ##
echo "dtoverlay=dwc2" >> boot/config.txt
sed -i "s/rootwait/rootwait modules-load=dwc2/g" boot/cmdline.txt

cp -a src/usb0.network      ${ROOT}/usr/lib/systemd/network/usb0.network
cp -a src/wlan0.network     ${ROOT}/usr/lib/systemd/network/wlan0.network
cp -a src/usbgadget         ${ROOT}/usr/sbin/usbgadget
cp -a src/usbgadget.service ${ROOT}/usr/lib/systemd/system/usbgadget.service 
## USB CDC ECM network device ##
enable_systemd network/usb0.network          network/usb0.network 
## serial port device ##
enable_systemd systemd/system/getty@.service system/multi-user.target.wants/getty@ttyGS0.service
## multi-gadget device ##
enable_systemd system/usbgadget.service      system/sysinit.target.wants/usbgadget.service

echo -e "\x1b[1;36mFinish create PwnPi\x1b[m"
