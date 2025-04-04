#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

echo "                                     ./////,                    "
echo "                                 ./////==//////*                "
echo "                                ////.  ___   ////.              "
echo "                         ,**,. ////  ,////A,  */// ,**,.        "
echo "                    ,/////////////*  */////*  *////////////A    "
echo "                   ////'        \VA.   '|'   .///'       '///*  "
echo "                  *///  .*///*,         |         .*//*,   ///* "
echo "                  (///  (//////)**--_./////_----*//////)   ///) "
echo "                   V///   '°°°°      (/////)      °°°°'   ////  "
echo "                    V/////(////////\. '°°°' ./////////(///(/'   "
echo "                       'V/(/////////////////////////////V'      "

echo "Preparing public-key..."

#mkdir -p /root/.ssh
#echo "${PUBLIC_KEY}" > /root/.ssh/id_rsa.pub
#echo "${PUBLIC_KEY}" > /root/.ssh/authorized_keys
#chown -R root:root /root/.ssh
#chmod -R 700 /root
#chmod -R 600 /root/.ssh/*

FQDN="$(cat /etc/ces/node_master)"
export FQDN

echo "Starting ssh"
/usr/sbin/sshd -e

echo "Starting exporter..."
/ces-exporter