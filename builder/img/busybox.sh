#!/bin/bash

TMP="$(mktemp --directory)"

apt-get update && apt-get install --yes libdigest-sha-perl && apt-get clean
URL="http://archive.ubuntu.com/ubuntu/pool/main/b/busybox/busybox-static_1.27.2-2ubuntu4_amd64.deb"
SHA="2d07c13235a3215991530f573b555900df06ae888c5f0299ae87d58f67caf5cd"
curl -fSLo "${TMP}/busybox.deb" "${URL}"
echo "${SHA}  ${TMP}/busybox.deb" | shasum -a "256" -c -

dpkg -i "${TMP}/busybox.deb"

mkdir "${TMP}/root"
cd "${TMP}/root"
mkdir bin etc dev dev/pts lib proc sys tmp
touch etc/resolv.conf
cp /etc/nsswitch.conf etc/nsswitch.conf
echo root:x:0:0:root:/:/bin/sh > etc/passwd
echo root:x:0: > etc/group
ln -s lib lib64
ln -s bin sbin
cp /bin/busybox bin
for name in $(busybox --list); do
  ln -s busybox "bin/${name}"
done
cp /lib/x86_64-linux-gnu/lib{c,dl,nsl,nss_*,pthread,resolv}.so.* lib
cp /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 lib

mksquashfs "${TMP}/root" "/mnt/out/layer.squashfs" -noappend
