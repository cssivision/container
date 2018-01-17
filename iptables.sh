#!/usr/bin/env bash

set -x

VETH="veth"
VETH_ADDR="10.88.37.11"

iptables -P FORWARD DROP
iptables -F FORWARD

# Flush nat rules.
iptables -t nat -F

# Enable masquerading of 10.200.1.0.
iptables -t nat -A POSTROUTING -s ${VETH_ADDR}/24 -o eth0 -j MASQUERADE

iptables -A FORWARD -i eth0 -o ${VETH} -j ACCEPT
iptables -A FORWARD -o eth0 -i ${VETH} -j ACCEPT