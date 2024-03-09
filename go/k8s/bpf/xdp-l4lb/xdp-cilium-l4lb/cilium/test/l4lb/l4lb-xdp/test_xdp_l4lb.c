

#include <stdint.h>
#include <stdbool.h>

#include <linux/bpf.h>
//#include <linux/stddef.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/icmp.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <linux/udp.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>





static __always_inline int process_packet(void *data, __u64 off, void *data_end, struct xdp_md *xdp)
{

}


SEC("xdp-test-v4")
int balancer_ingress_v4(struct xdp_md *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct ethhdr *eth = data;
    __u32 nh_off;

    nh_off = sizeof(struct ethhdr);
    if (data + nh_off > data_end)
        return XDP_DROP;
    if (eth->h_proto == bpf_htons(ETH_P_IP)) // ipv4
        return process_packet(data, nh_off, data_end, 0, ctx);
    else
        return XDP_DROP;
}


char _license[] SEC("license") = "GPL";
