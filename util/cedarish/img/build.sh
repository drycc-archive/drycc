#!/bin/bash

# Derived from https://github.com/heroku/stack-images/blob/master/bin/cedar-14.sh
source /etc/lsb-release

echo deb http://archive.ubuntu.com/ubuntu "${DISTRIB_CODENAME}" main restricted >/etc/apt/sources.list
echo deb http://archive.ubuntu.com/ubuntu "${DISTRIB_CODENAME}"-updates main restricted >>/etc/apt/sources.list
echo deb http://archive.ubuntu.com/ubuntu "${DISTRIB_CODENAME}" universe >>/etc/apt/sources.list
echo deb http://archive.ubuntu.com/ubuntu "${DISTRIB_CODENAME}"-updates universe >>/etc/apt/sources.list
echo deb http://archive.ubuntu.com/ubuntu "${DISTRIB_CODENAME}"-security main restricted >>/etc/apt/sources.list
echo deb http://archive.ubuntu.com/ubuntu "${DISTRIB_CODENAME}"-security universe >>/etc/apt/sources.list

apt-get update
apt-get dist-upgrade -y

apt-get install -y \
  autoconf \
  bind9-host \
  bison \
  build-essential \
  coreutils \
  curl \
  daemontools \
  dnsutils \
  ed \
  git \
  imagemagick \
  iputils-tracepath \
  libbz2-dev \
  libcurl4-openssl-dev \
  libevent-dev \
  libglib2.0-dev \
  libjpeg-dev \
  libmagickwand-dev \
  libmysqlclient-dev \
  libncurses5-dev \
  libpq-dev \
  libpq5 \
  libreadline6-dev \
  libssl-dev \
  libxml2-dev \
  libxslt-dev \
  netcat-openbsd \
  openjdk-8-jdk \
  openjdk-8-jre-headless \
  openssh-client \
  openssh-server \
  postgresql-server-dev-all \
  python \
  python-dev \
  ruby \
  ruby-dev \
  socat \
  stunnel \
  syslinux \
  tar \
  telnet \
  zip \
  zlib1g-dev \
  pigz

# Install locales
apt-cache search language-pack \
  | cut -d ' ' -f 1 \
  | grep -v '^language\-pack\-\(touch\|gnome\|kde\)\-' \
  | grep -v '\-base$' \
  | xargs apt-get install -y --no-install-recommends

rm -rf /var/cache/apt/archives/*.deb
rm -rf /root/*
rm -rf /tmp/*
rm /etc/ssh/ssh_host_*
