#!/usr/bin/env bash

set -x

DEVICE_NAME="container0"
DEVICE_ADDR="10.88.37.1"

iptables -P FORWARD DROP
iptables -F FORWARD

# Flush nat rules.
iptables -t nat -F

# Enable masquerading of ${DEVICE_ADDR}.
iptables -t nat -A POSTROUTING -s ${DEVICE_ADDR}/24 -o eth0 -j MASQUERADE

iptables -A FORWARD -i eth0 -o ${DEVICE_NAME} -j ACCEPT
iptables -A FORWARD -o eth0 -i ${DEVICE_NAME} -j ACCEPT