





add_executable(xdp_cilium_l4lb_test
        cilium/test/xdp-l4lb/slb.bpf.c
        cilium/test/tcp/tcprtt/demo2/tcprtt_sockops.c
        cilium/test/linux/linux/pkt_cls.h
        cilium/test/linux/bpf/bpf_endian.h
        cilium/test/tcp/tcpsyncookie/tcp_syncookie.c
        cilium/test/linux/bpf/bpf_helpers.h
        cilium/test/linux/bpf/bpf_tracing.h
        cilium/test/linux/linux/if_ether.h
        cilium/test/linux/linux/ip.h
)
