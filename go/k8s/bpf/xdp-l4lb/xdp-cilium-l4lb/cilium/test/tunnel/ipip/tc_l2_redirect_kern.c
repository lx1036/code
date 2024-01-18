

/**
 * /root/linux-5.10.142/samples/bpf/tc_l2_redirect_kern.c
 * /root/linux-5.10.142/net/ipv4/ipip.c
 */

#include <stdbool.h>
#include <string.h>

#include <linux/stddef.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/mpls.h>
#include <linux/tcp.h>
// #include <linux/udp.h>
#include <linux/pkt_cls.h>
#include <linux/types.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>


#include <linux/if_packet.h>
#include <linux/filter.h>
// #include <net/ipv6.h>


#define _htonl __builtin_bswap32
// #define SEC(NAME) __attribute__((section(NAME), used))
#define PIN_GLOBAL_NS		2

struct bpf_elf_map {
	__u32 type;
	__u32 size_key;
	__u32 size_value;
	__u32 max_elem;
	__u32 flags;
	__u32 id;
	__u32 pinning;
	__u32 inner_id;
	__u32 inner_idx;
};


struct bpf_elf_map SEC("maps") tun_iface = {
	.type = BPF_MAP_TYPE_ARRAY,
	.size_key = sizeof(int),
	.size_value = sizeof(int),
	.pinning = PIN_GLOBAL_NS,
	.max_elem = 1,
};


// 目标 vip 在网段 10.10.1.0/24 内
static __always_inline bool is_vip_addr(__be16 eth_proto, __be32 daddr)
{
	if (eth_proto == __bpf_htons(ETH_P_IP))
		return (_htonl(0xffffff00) & daddr) == _htonl(0x0a0a0100); // 10.10.1.0/24
	else if (eth_proto == __bpf_htons(ETH_P_IPV6))
		return (daddr == _htonl(0x2401face));

	return false;
}

// vens2 tc ingress 主要拦截 10.10.1.0/24 的包, 此时的包为 ipip，所以不会被丢弃:
// Outer 10.2.1.1(ve2) > 10.2.1.102(vens2) Inner 10.1.1.101(vens1) > 10.10.1.102
SEC("drop_non_tun_vip")
int _drop_non_tun_vip(struct __sk_buff *skb)
{
	struct bpf_tunnel_key tkey = {};
	void *data = (void *)(long)skb->data;
	struct ethhdr *eth = data;
	void *data_end = (void *)(long)skb->data_end;

	if (data + sizeof(*eth) > data_end)
		return TC_ACT_OK;

	if (eth->h_proto == __bpf_htons(ETH_P_IP)) {
		struct iphdr *iph = data + sizeof(*eth);

		if (data + sizeof(*eth) + sizeof(*iph) > data_end)
			return TC_ACT_OK;

		// 如果访问 ip 在网段 10.10.1.0/24 内, 则丢弃包
		if (is_vip_addr(eth->h_proto, iph->daddr))
			return TC_ACT_SHOT;
	} else if (eth->h_proto == __bpf_htons(ETH_P_IPV6)) {
		struct ipv6hdr *ip6h = data + sizeof(*eth);

		if (data + sizeof(*eth) + sizeof(*ip6h) > data_end)
			return TC_ACT_OK;

		if (is_vip_addr(eth->h_proto, ip6h->daddr.s6_addr32[0]))
			return TC_ACT_SHOT;
	}

	return TC_ACT_OK;
}

// 把回包转发到 veth2->tun1.ipip
SEC("l2_to_iptun_ingress_forward")
int _l2_to_iptun_ingress_forward(struct __sk_buff *skb)
{
	struct bpf_tunnel_key tkey = {};
	void *data = (void *)(long)skb->data;
	struct ethhdr *eth = data;
	void *data_end = (void *)(long)skb->data_end;
	int key = 0, *ifindex;

	int ret;

	if (data + sizeof(*eth) > data_end)
		return TC_ACT_OK;

	ifindex = bpf_map_lookup_elem(&tun_iface, &key);
	if (!ifindex)
		return TC_ACT_OK;

	if (eth->h_proto == __bpf_htons(ETH_P_IP)) {
		char fmt4[] = "ingress forward to ifindex:%d daddr4:%x\n"; // "ingress forward to ifindex:8 daddr4:a020101"
		struct iphdr *iph = data + sizeof(*eth);

		if (data + sizeof(*eth) + sizeof(*iph) > data_end)
			return TC_ACT_OK;

		// ve2 收到的包已经是封装后的 ipip 包
		if (iph->protocol != IPPROTO_IPIP)
			return TC_ACT_OK;

        // cat /sys/kernel/debug/tracing/trace_pipe
        // cat /sys/kernel/debug/tracing/trace
        // 此时的回包为 10.2.1.102 > 10.2.1.1(10.10.1.102 > 10.1.1.101)
		bpf_trace_printk(fmt4, sizeof(fmt4), *ifindex, _htonl(iph->daddr)); // __u32 -> a020101=10.2.1.1(ve2, host)

		// BPF_F_INGRESS 表示 redirect 到 ifindex 的 ingress 这个 hook
		return bpf_redirect(*ifindex, BPF_F_INGRESS);
	} else if (eth->h_proto == __bpf_htons(ETH_P_IPV6)) {
		char fmt6[] = "ingress forward to ifindex:%d daddr6:%x::%x\n";
		struct ipv6hdr *ip6h = data + sizeof(*eth);

		if (data + sizeof(*eth) + sizeof(*ip6h) > data_end)
			return TC_ACT_OK;

		if (ip6h->nexthdr != IPPROTO_IPIP &&
		    ip6h->nexthdr != IPPROTO_IPV6)
			return TC_ACT_OK;

		bpf_trace_printk(fmt6, sizeof(fmt6), *ifindex,
				 _htonl(ip6h->daddr.s6_addr32[0]),
				 _htonl(ip6h->daddr.s6_addr32[3]));
		return bpf_redirect(*ifindex, BPF_F_INGRESS);
	}

	return TC_ACT_OK;
}


SEC("l2_to_iptun_ingress_redirect")
int _l2_to_iptun_ingress_redirect(struct __sk_buff *skb)
{
	struct bpf_tunnel_key tkey = {};
	void *data = (void *)(long)skb->data;
	struct ethhdr *eth = data;
	void *data_end = (void *)(long)skb->data_end;
	int key = 0, *ifindex;

	int ret;

	if (data + sizeof(*eth) > data_end)
		return TC_ACT_OK;

	ifindex = bpf_map_lookup_elem(&tun_iface, &key);
	if (!ifindex)
		return TC_ACT_OK;

	if (eth->h_proto == __bpf_htons(ETH_P_IP)) {
		char fmt4[] = "e/ingress redirect daddr4:%x to ifindex:%d\n"; // 去的 icmp 包: 10.1.1.101 > 10.10.1.102
		struct iphdr *iph = data + sizeof(*eth);
		__be32 daddr = iph->daddr;

		if (data + sizeof(*eth) + sizeof(*iph) > data_end)
			return TC_ACT_OK;

		if (!is_vip_addr(eth->h_proto, daddr)) // 10.1.1.101 > 10.10.1.102
//		char fmt_is_vip_addr[] = "[fmt_is_vip_addr]ingress redirect daddr4:%x to ifindex:%d\n";
//		if (!is_vip_addr(eth->h_proto, daddr))
//		    bpf_trace_printk(fmt_is_vip_addr, sizeof(fmt_is_vip_addr), _htonl(daddr), *ifindex);
			return TC_ACT_OK;

		bpf_trace_printk(fmt4, sizeof(fmt4), _htonl(daddr), *ifindex);
	} else {
		return TC_ACT_OK;
	}

    // 回的包 ipip: 10.2.1.102(10.10.1.102) > 10.2.1.1(10.1.1.101)
	tkey.tunnel_id = 10000;
	tkey.tunnel_ttl = 64;
	tkey.remote_ipv4 = 0x0a010265; /* 10.1.2.101 eth0(ns2) 网卡地址 */
	// Populate tunnel metadata for packet associated to *skb.*
	bpf_skb_set_tunnel_key(skb, &tkey, sizeof(tkey), 0);
	return bpf_redirect(*ifindex, 0); // egress direction
}

/*
TC_ACT_OK: accept packet
TC_ACT_SHOT: drop packet
*/

char _license[] SEC("license") = "GPL";
