



// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_xdp_vlan.c

/**
 * https://info.support.huawei.com/info-finder/encyclopedia/zh/VLAN.html
 * https://forum.huawei.com/enterprise/zh/thread/580889430837837824
 * https://datatracker.ietf.org/doc/html/rfc2674
 * https://wiki.wireshark.org/VLAN
 * https://en.wikipedia.org/wiki/VLAN
 * https://wiki.wireshark.org/CaptureSetup/VLAN
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
    __be16 h_vlan_TCI; // TCI(VLAN tag control information)
    __be16 h_vlan_encapsulated_proto;
};

struct parse_pkt {
    __u16 l3_proto;
    __u16 l3_offset;
    __u16 vlan_outer;
    __u16 vlan_inner;
    __u8  vlan_outer_offset;
    __u8  vlan_inner_offset;
};

static __always_inline
bool parse_eth_frame(struct ethhdr *eth, void *data_end, struct parse_pkt *pkt) {
    __u16 eth_type;
    __u8 offset;

    offset = sizeof(*eth);
    /* Make sure packet is large enough for parsing eth + 2 VLAN headers */
    if ((void *)eth + offset + (2*sizeof(struct vlan_hdr)) > data_end)
        return false;

    eth_type = eth->h_proto;
    /* Handle outer VLAN tag */
    if (eth_type == bpf_htons(ETH_P_8021Q) || eth_type == bpf_htons(ETH_P_8021AD)) {
        struct vlan_hdr *vlan_hdr;
        vlan_hdr = (void *)eth + offset;
        pkt->vlan_outer_offset = offset;
        pkt->vlan_outer = bpf_ntohs(vlan_hdr->h_vlan_TCI) & VLAN_VID_MASK; // 去除前 4bits
        eth_type        = vlan_hdr->h_vlan_encapsulated_proto;
        offset += sizeof(*vlan_hdr);
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

/* Changing VLAN to zero, have same practical effect as removing the VLAN. */
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
        vlan_hdr->h_vlan_TCI = bpf_htons((bpf_ntohs(vlan_hdr->h_vlan_TCI) & 0xf000) | TO_VLAN); // 只保留前 4bits
    }

    return XDP_PASS;
}

/*=====================================
 *  BELOW: TC-hook based ebpf programs
 * ====================================
 * The TC-clsact eBPF programs (currently) need to be attach via TC commands
 */
/*
Commands to setup TC to use above bpf prog:

export ROOTDEV=ixgbe2
export FILE=xdp_vlan01_kern.o

# Re-attach clsact to clear/flush existing role
tc qdisc del dev $ROOTDEV clsact 2> /dev/null ;\
tc qdisc add dev $ROOTDEV clsact

# Attach BPF prog EGRESS
tc filter add dev $ROOTDEV egress prio 1 handle 1 bpf da obj $FILE sec tc_vlan_push
tc filter show dev $ROOTDEV egress
*/
SEC("tc_vlan_push")
int tc_progA(struct __sk_buff *ctx) {
    // 增加一个 vlan_tci 并更新 checksum
    bpf_skb_vlan_push(ctx, bpf_htons(ETH_P_8021Q), TESTVLAN);
    return TC_ACT_OK;
}


char _license[] SEC("license") = "GPL";
