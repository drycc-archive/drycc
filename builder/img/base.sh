#!/bin/bash

URL="https://partner-images.canonical.com/core/xenial/current/"
TAR="ubuntu-xenial-core-cloudimg-amd64-root.tar.gz"
TMP="$(mktemp --directory)"

SHA="$(curl -s "${URL}""/SHA256SUMS"|grep "${TAR}"|awk '{print $1}')"

curl -fSLo "${TMP}/ubuntu.tar.gz" "${URL}"/"${TAR}"
echo "${SHA}  ${TMP}/ubuntu.tar.gz" | sha256sum -c -

mkdir -p "${TMP}/root"
tar xf "${TMP}/ubuntu.tar.gz" -C "${TMP}/root"

cp "/etc/resolv.conf" "${TMP}/root/etc/resolv.conf"
mount --bind "/dev/pts" "${TMP}/root/dev/pts"
cleanup() {
  umount "${TMP}/root/dev/pts"
  >"${TMP}/root/etc/resolv.conf"
}
trap cleanup EXIT

chroot "${TMP}/root" bash -e < "builder/ubuntu-setup.sh"

mksquashfs "${TMP}/root" "/mnt/out/layer.squashfs" -noappend
