
#!/usr/bin/env bash


# https://github.com/P4-Research/ebpf-demos/blob/master/arp-resp/install_xdp.sh

clang -Wno-unused-value -Wno-pointer-sign -Wno-compare-distinct-pointer-types -Wno-gnu-variable-sized-type-not-at-end -Wno-tautological-compare -O2 -emit-llvm -g -c arp-resp.c -o -| llc -march=bpf -filetype=obj -o dropper.o
sudo ip link set dev s1-eth1 xdp off
sudo ip link set dev s1-eth1 xdp obj dropper.o sec xdp_arp_resp
