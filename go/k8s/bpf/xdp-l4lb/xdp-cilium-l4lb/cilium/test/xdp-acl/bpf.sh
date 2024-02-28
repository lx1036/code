#!/bin/bash

# 本地运行时，不会走 eth0-acl 网卡，而是 lo 网卡 127.0.0.1.9091 > 127.0.0.1.9090

ip link add dev eth0-acl type dummy
ip link set dev eth0-acl up
ip addr add 10.20.30.40/24 dev eth0-acl
ip addr add 10.20.30.41/24 dev eth0-acl
