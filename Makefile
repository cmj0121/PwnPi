ARCH_SITE_LINK=http://os.archlinuxarm.org/os/
RPI3_IMG=ArchLinuxARM-rpi-2-latest.tar.gz
RPI0w_IMG=ArchLinuxARM-rpi-latest.tar.gz

HOSTNAME=PwnPi

SUBDIR=plugins


.PHONY: install_rpi3

install_rpi0w: rpi0w
	@mount $(DEV)1 boot
	@mount $(DEV)2 root
	trap 'umount {root,boot}' EXIT; \
		echo "dtoverlay=dwc2" >> boot/config.txt && \
		sed -i "s/rootwait/rootwait modules-load=dwc2/g" boot/cmdline.txt && \
		cp -ar src/*.network root/usr/lib/systemd/network/ && \
		cp -ar src/*.service root/usr/lib/systemd/system/  && \
		cp -ar src/usbgadget root/usr/sbin && \
		ln -s /usr/lib/systemd/network/usb0.network     root/etc/systemd/network/usb0.network && \
		ln -s /usr/lib/systemd/system/getty@.service    root/etc/systemd/system/multi-user.target.wants/getty@ttyGS0.service && \
		ln -s /usr/lib/systemd/system/usbgadget.service root/etc/systemd/system/sysinit.target.wants/usbgadget.service

install_rpi3: rpi3
	@mount $(DEV)1 boot
	@mount $(DEV)2 root
	trap 'umount {root,boot}' EXIT; \
		echo "dtoverlay=dwc2" >> boot/config.txt && \
		sed -i "s/rootwait/rootwait modules-load=dwc2/g" boot/cmdline.txt && \
		cp -ar src/*.network root/usr/lib/systemd/network/ && \
		cp -ar src/*.service root/usr/lib/systemd/system/  && \
		cp -ar src/usbgadget root/usr/sbin && \
		ln -s /usr/lib/systemd/network/usb0.network     root/etc/systemd/network/usb0.network && \
		ln -s /usr/lib/systemd/system/getty@.service    root/etc/systemd/system/multi-user.target.wants/getty@ttyGS0.service && \
		ln -s /usr/lib/systemd/system/usbgadget.service root/etc/systemd/system/sysinit.target.wants/usbgadget.service

rpi0w: /tmp/$(RPI0w_IMG) format_sd
	@mkdir -p {root,boot}
	@mount $(DEV)1 boot
	@mount $(DEV)2 root
	trap 'umount {boot,root}' EXIT; bsdtar -xpf /tmp/$(RPI0w_IMG) -C root && sync && mv ./root/boot/* ./boot

rpi3: /tmp/$(RPI3_IMG) format_sd
	@mkdir -p {root,boot}
	@mount $(DEV)1 boot
	@mount $(DEV)2 root
	trap 'umount {boot,root}' EXIT; bsdtar -xpf /tmp/$(RPI3_IMG) -C root && sync && mv ./root/boot/* ./boot


.PHONY: rpi3 root_check

root_check:
	@[ `id -u` == 0 ]  || (echo "Need root permission" && exit -1)

format_sd: root_check
	@[ -b "$(DEV)" ] || (echo "DEV \`$(DEV)\` is not valid device" && exit -1)
	parted -s $(DEV) mklabel msdos
	parted -s $(DEV) mkpart primary fat32 1 128
	parted -s $(DEV) mkpart primary ext4 --  128 -1
	mkfs.vfat $(DEV)1
	mkfs.ext4 -F $(DEV)2

/tmp/$(RPI3_IMG):
	wget $(ARCH_SITE_LINK)/$(RPI3_IMG) -O /tmp/$(RPI3_IMG)
	wget $(ARCH_SITE_LINK)/$(RPI3_IMG).md5 -qO - | sed "s%$(RPI3_IMG)%/tmp/$(RPI3_IMG)%g" | md5sum -c

.PHONY: clean

clean: $(SUBDIR)

