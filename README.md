# PwnPi

This project is the PoC (Proof-of-Concept) project that raspberry pi can be used as pwn tool,
and using the [RPi zero w][0] as the main development hardware.

The base image is based on the [official][1] and install the extra tools for pwn.

## Install

### Install Packages

Before install the package you need to setup your SSH key to the raspberry pi. It makes you easy
to connect to the raspberry pi without typing the password. You can just run `make prologue` to
setup your SSH key to your PwnPi.

[0]: https://www.raspberrypi.org/products/raspberry-pi-zero-w/
[1]: https://www.raspberrypi.com/software/operating-systems/
