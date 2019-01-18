#!/bin/bash

set -eo pipefail

version="1.11"
shasum="8ffb718cab4371df3495d8f6c0b15c7e6dc28fea"
dir="/usr/local"

apt-get update
apt-get install --yes git build-essential unzip
apt-get clean

curl -fsSLo /tmp/go.tar.gz "https://dl.google.com/go/go${version}.linux-amd64.tar.gz"
echo "${shasum}  /tmp/go.tar.gz" | shasum -c -
tar xzf /tmp/go.tar.gz -C "${dir}"
rm /tmp/go.tar.gz

export GOROOT="/usr/local/go"
export GOPATH="/go"
export PATH="${GOROOT}/bin:${PATH}"

# install protobuf compiler
tmpdir=$(mktemp --directory)
trap "rm -rf ${tmpdir}" EXIT
curl -sL https://github.com/google/protobuf/releases/download/v3.3.0/protoc-3.3.0-linux-x86_64.zip > "${tmpdir}/protoc.zip"
unzip -d "${tmpdir}/protoc" "${tmpdir}/protoc.zip"
mv "${tmpdir}/protoc" /opt
ln -s /opt/protoc/bin/protoc /usr/local/bin/protoc

mkdir -p "${GOPATH}/src/github.com/drycc"
ln -nfs "$(pwd)" "${GOPATH}/src/github.com/drycc/drycc"

cp "builder/go-wrapper.sh" "/usr/local/bin/go"
cp "builder/go-wrapper.sh" "/usr/local/bin/cgo"
