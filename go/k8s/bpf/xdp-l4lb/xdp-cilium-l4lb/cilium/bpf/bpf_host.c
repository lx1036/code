





#include <lib/maps.h>

#include <linux/if_ether.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_endian.h>

#include <lib/host_firewall.h>



SEC("to-host")
int to_host(struct __sk_buff *ctx) {
    __u16 proto = 0;
    int ret = TC_ACT_OK;

    void *data_end = (void *)(long)(ctx->data_end);
    void *data = (void *)(long)(ctx->data);
    struct ethhdr *eth = data;
    if (data + ETH_HLEN > data_end) {
        return TC_ACT_SHOT;
    }

    proto = eth->h_proto;
    switch (proto) {
    case bpf_htons(ETH_P_ARP):
        ret = TC_ACT_OK;
        break;

    case bpf_htons(ETH_P_IP):
        ret = ipv4_host_policy_ingress(ctx, &src_id, &trace);
        break;

    default:
        ret = TC_ACT_SHOT;
        break;
    }

    return ret;
}
