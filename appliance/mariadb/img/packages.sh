#!/bin/bash

source /etc/lsb-release
export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get install -y software-properties-common apt-transport-https
apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 \
  199369E5404BD5FC7D2FE43BCBCB082A1BB943DB \
  4D1BB29D63D98E422B2113B19334A25F8507EFA5

echo deb http://sfo1.mirrors.digitalocean.com/mariadb/repo/10.1/ubuntu "${DISTRIB_CODENAME}" main >>/etc/apt/sources.list
echo deb http://repo.percona.com/apt "${DISTRIB_CODENAME}" main >>/etc/apt/sources.list
apt-get update
apt-get install -y sudo
apt-get install -y --allow-unauthenticated mariadb-server percona-xtrabackup
apt-get clean
apt-get autoremove -y
