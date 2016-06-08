#!/bin/bash

ID=1
iwpan phy phy0 set channel 0 11
ip link add link wpan0 name lowpan0 type lowpan
iwpan dev wpan0 set pan_id 0xbeef
ip link set wpan0 up
ip link set lowpan0 up
ip address add dev lowpan0 beef::$ID/64