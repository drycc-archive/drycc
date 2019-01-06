# Packer Templates

This directory contains Packer templates for building machine images that
represent an Ubuntu target system for Drycc. These are essentially stock images
of Ubuntu 16.04 with Drycc installed.

## Usage

First, [install Packer](http://www.packer.io/intro/getting-started/setup.html).
Then, clone this repository and `cd` into the `util/packer` target directory.

## Vagrant Template

Currently supports:
 * VirtualBox
 * VMWare Fusion

To build just a VirtualBox image for use with the Drycc Vagrantfile:

```
$ packer build -only=virtualbox-iso -var-file ubuntu-xenial.json ubuntu.json
```

## Then What?

At the end of any of these, you'll have a snapshot or image ready to go.
