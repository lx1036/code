



// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_xdp_vlan.c

/**
 * https://info.support.huawei.com/info-finder/encyclopedia/zh/VLAN.html
 * https://forum.huawei.com/enterprise/zh/thread/580889430837837824
 * https://datatracker.ietf.org/doc/html/rfc2674
 * https://wiki.wireshark.org/VLAN
 * https://en.wikipedia.org/wiki/VLAN
 * https://wiki.wireshark.org/CaptureSetup/VLAN
 *
 * 抓包:
 * ip netns exec ns1 ping -c 2 100.64.41.2
   ip netns exec ns2 ping -c 2 100.64.41.1

 * ip netns exec ns1 tcpdump -i veth1 -nneevv -A icmp -w ns1_veth1.pcap (这里 vlan_id 为 0)
 *
 * ip netns exec ns2 tcpdump -i veth2 -nneevv -A icmp -w ns2_veth2.pcap (这里抓包看到 vlan_id 4011)
 * ip netns exec ns2 tcpdump -i veth2.4011 -nneevv -A icmp -w ns2_veth2.4011.pcap (这里看不到 vlan 头)
 */


#include <stddef.h>
#include <stdbool.h>
#include <string.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/if_vlan.h>
#include <linux/in.h>
#include <linux/pkt_cls.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>



/* Hint, VLANs are choosen to hit network-byte-order issues */
#define TESTVLAN 4011 /* 0xFAB */
// #define TO_VLAN  4000 /* 0xFA0 (hint 0xOA0 = 160) */

#define TO_VLAN	0

#define VLAN_VID_MASK		0x0fff /* VLAN Identifier */

/* linux/if_vlan.h have not exposed this as UAPI, thus mirror some here
 *
 *	struct vlan_hdr - vlan header
 *	@h_vlan_TCI: priority and VLAN ID
 *	@h_vlan_encapsulated_proto: packet type ID or len
 */
struct vlan_hdr { // 4字节
    __be16 h_vlan_TCI; // TCI(VLAN tag control information), Priority(3)+DEI(1)+ID(12)
    __be16 h_vlan_encapsulated_proto; // ipv4 0x0800
};

struct parse_pkt {
    __u16 l3_proto;
    __u16 l3_offset;
    __u16 vlan_outer;
    __u16 vlan_inner;
    __u8  vlan_outer_offset;
    __u8  vlan_inner_offset;
};

// 协议顺序: ethhdr->ETH_P_8021Q->iphdr->icmphdr
static __always_inline
bool parse_eth_frame(struct ethhdr *eth, void *data_end, struct parse_pkt *pkt) {
    __u16 eth_type;
    __u8 offset;

    offset = sizeof(*eth);
    /* Make sure packet is large enough for parsing eth + 2 VLAN headers */
    // 这里为何需要 2 个 vlan_hdr
    if ((void *)eth + offset + (2*sizeof(struct vlan_hdr)) > data_end)
        return false;

    eth_type = eth->h_proto; // Type: 802.1Q Virtual LAN (0x8100)
    /* Handle outer VLAN tag */
    if (eth_type == bpf_htons(ETH_P_8021Q) || eth_type == bpf_htons(ETH_P_8021AD)) {
        struct vlan_hdr *vlan_hdr;
        vlan_hdr = (void *)eth + offset;
        pkt->vlan_outer_offset = offset; // ethhdr 的长度，vlan_outer 指针
        pkt->vlan_outer = bpf_ntohs(vlan_hdr->h_vlan_TCI) & VLAN_VID_MASK; // 去除前 4bits, 获取 vlan_id
        eth_type        = vlan_hdr->h_vlan_encapsulated_proto; // Type: IPv4 (0x0800)
        offset += sizeof(*vlan_hdr); // vlan_inner 指针
    }

    /* Handle inner (double) VLAN tag */
    if (eth_type == bpf_htons(ETH_P_8021Q) || eth_type == bpf_htons(ETH_P_8021AD)) {
        struct vlan_hdr *vlan_hdr;
        vlan_hdr = (void *)eth + offset; // 第二个 vlan_hdr
        pkt->vlan_inner_offset = offset;
        pkt->vlan_inner = bpf_ntohs(vlan_hdr->h_vlan_TCI) & VLAN_VID_MASK;
        eth_type        = vlan_hdr->h_vlan_encapsulated_proto;
        offset += sizeof(*vlan_hdr);
    }

    pkt->l3_proto = bpf_ntohs(eth_type); /* Convert to host-byte-order */
    pkt->l3_offset = offset;

    return true;
}

// vlan_id 置0，等于 remove vlan
SEC("xdp_vlan_change")
int  xdp_prognum1(struct xdp_md *ctx)
{
    void *data_end = (void *)(long)ctx->data_end;
    void *data     = (void *)(long)ctx->data;
    struct parse_pkt pkt = { 0 };

    if (!parse_eth_frame(data, data_end, &pkt))
        return XDP_ABORTED;

    /* Change specific VLAN ID */
    if (pkt.vlan_outer == TESTVLAN) {
        struct vlan_hdr *vlan_hdr = data + pkt.vlan_outer_offset;
        /* Modifying VLAN, preserve top 4 bits */
        // bpf_ntohs: __be16->0x, bpf_htons: 0x->__be16, 等同于类型转换 hex<->int
        // 只保留前 4bits，因为 ns2 过来的包是一个 vlan.4011 包，然后经过 xdp 处理来解包 vlan，解包方式为把 ID(12) 字段置0
        vlan_hdr->h_vlan_TCI = bpf_htons((bpf_ntohs(vlan_hdr->h_vlan_TCI) & 0xf000) | TO_VLAN);
    }

    return XDP_PASS;
}

/*
 * Show XDP+TC can cooperate, on creating a VLAN rewriter.
 * 1. Create a XDP prog that can "pop"/remove a VLAN header.
 * 2. Create a TC-bpf prog that egress can add a VLAN header.
 */
SEC("xdp_vlan_remove_outer")
int  xdp_prognum2(struct xdp_md *ctx) {
    void *data_end = (void *) (long) ctx->data_end;
    void *data = (void *) (long) ctx->data;
    struct parse_pkt pkt = {0}; // 初始化全部字段置0
    char *dest;

    if (!parse_eth_frame(data, data_end, &pkt))
        return XDP_ABORTED;

    /* Skip packet if no outer VLAN was detected */
    if (pkt.vlan_outer_offset == 0)
        return XDP_PASS;




}


/**
# Attach BPF prog EGRESS
tc qdisc add dev $ROOTDEV clsact
tc qdisc del dev $ROOTDEV clsact
tc filter add dev $ROOTDEV egress prio 1 handle 1 bpf da obj $FILE sec tc_vlan_push
tc filter show dev $ROOTDEV egress
*/
SEC("tc_vlan_push")
int tc_progA(struct __sk_buff *ctx) {
    // 增加一个 vlan_tci 并更新 checksum，在 tc egress 增加一个 vlan hdr
    bpf_skb_vlan_push(ctx, bpf_htons(ETH_P_8021Q), TESTVLAN);
    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";


/**
 * 解释:
 *
 * ns1/veth1[100.64.41.1] ------> ns2/(veth2, veth2.4011[100.64.41.2]), veth1/veth2 又是一对 veth-pair
 * bpf, 挂载在 ns1/veth1:
 * packet --> xdp/xdp_vlan_change --> veth1 --> tc_egress/tc_vlan_push
 *
 */

