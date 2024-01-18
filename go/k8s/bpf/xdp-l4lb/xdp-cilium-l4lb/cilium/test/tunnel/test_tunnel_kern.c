

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tunnel_kern.c
 *
 */

/*
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of version 2 of the GNU General Public
 * License as published by the Free Software Foundation.
 */
#include <stdio.h>
#include <stddef.h>
#include <string.h>
#include <arpa/inet.h>
// /root/linux-5.10.142/include/uapi/linux/bpf.h
#include <linux/bpf.h>
// /root/linux-5.10.142/include/uapi/linux/if_ether.h
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/types.h>
#include <linux/socket.h>
#include <linux/pkt_cls.h>
#include <linux/erspan.h>
// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ERROR(ret) do {\
		char fmt[] = "ERROR line:%d ret:%d\n";\
		bpf_trace_printk(fmt, sizeof(fmt), __LINE__, ret); \
	} while (0)

struct geneve_opt {
	__be16	opt_class;
	__u8	type;
	__u8	length:5;
	__u8	r3:1;
	__u8	r2:1;
	__u8	r1:1;
	__u8	opt_data[8]; /* hard-coded to 8 byte */
};

struct vxlan_metadata {
	__u32     gbp;
};

// 定义一个函数，输入是__u32类型IP地址，输出是一个已分配内存的字符串
static __always_inline char* u32toIpStr(__u32 ip) {
    static char str[16]; // IP地址字符串长度最大为"255.255.255.255"即15个字符加结束符'\0'

    // 分离IP地址的四个字节
    unsigned char bytes[4];
    bytes[0] = (ip >> 24) & 0xFF;
    bytes[1] = (ip >> 16) & 0xFF;
    bytes[2] = (ip >> 8) & 0xFF;
    bytes[3] = ip & 0xFF;

    // 将字节转换为点分十进制字符串
    sprintf(str, "%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3]);

    return str;
}

SEC("gre_set_tunnel")
int _gre_set_tunnel(struct __sk_buff *skb)
{
	int ret;
	struct bpf_tunnel_key key;

	__builtin_memset(&key, 0x0, sizeof(key));
	key.remote_ipv4 = 0xac100164; /* 172.16.1.100 */
	key.tunnel_id = 2;
	key.tunnel_tos = 0;
	key.tunnel_ttl = 64;

	ret = bpf_skb_set_tunnel_key(skb, &key, sizeof(key),
				     BPF_F_ZERO_CSUM_TX | BPF_F_SEQ_NUMBER);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	return TC_ACT_OK;
}

SEC("gre_get_tunnel")
int _gre_get_tunnel(struct __sk_buff *skb)
{
	int ret;
	struct bpf_tunnel_key key;
	char fmt[] = "key %d remote ip 0x%x\n";

	ret = bpf_skb_get_tunnel_key(skb, &key, sizeof(key), 0);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	bpf_trace_printk(fmt, sizeof(fmt), key.tunnel_id, key.remote_ipv4);
	return TC_ACT_OK;
}

SEC("erspan_set_tunnel")
int _erspan_set_tunnel(struct __sk_buff *skb)
{
	struct bpf_tunnel_key key;
	struct erspan_metadata md;
	int ret;

	__builtin_memset(&key, 0x0, sizeof(key));
	key.remote_ipv4 = 0xac100164; /* 172.16.1.100 */
	key.tunnel_id = 2;
	key.tunnel_tos = 0;
	key.tunnel_ttl = 64;

	ret = bpf_skb_set_tunnel_key(skb, &key, sizeof(key),
				     BPF_F_ZERO_CSUM_TX);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	__builtin_memset(&md, 0, sizeof(md));
#ifdef ERSPAN_V1
	md.version = 1;
	md.u.index = bpf_htonl(123);
#else
	__u8 direction = 1;
	__u8 hwid = 7;

	md.version = 2;
	md.u.md2.dir = direction;
	md.u.md2.hwid = hwid & 0xf;
	md.u.md2.hwid_upper = (hwid >> 4) & 0x3;
#endif

	ret = bpf_skb_set_tunnel_opt(skb, &md, sizeof(md));
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	return TC_ACT_OK;
}

SEC("erspan_get_tunnel")
int _erspan_get_tunnel(struct __sk_buff *skb)
{
	char fmt[] = "key %d remote ip 0x%x erspan version %d\n";
	struct bpf_tunnel_key key;
	struct erspan_metadata md;
	__u32 index;
	int ret;

	ret = bpf_skb_get_tunnel_key(skb, &key, sizeof(key), 0);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	ret = bpf_skb_get_tunnel_opt(skb, &md, sizeof(md));
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	bpf_trace_printk(fmt, sizeof(fmt),
			key.tunnel_id, key.remote_ipv4, md.version);

#ifdef ERSPAN_V1
	char fmt2[] = "\tindex %x\n";

	index = bpf_ntohl(md.u.index);
	bpf_trace_printk(fmt2, sizeof(fmt2), index);
#else
	char fmt2[] = "\tdirection %d hwid %x timestamp %u\n";

	bpf_trace_printk(fmt2, sizeof(fmt2),
			 md.u.md2.dir,
			 (md.u.md2.hwid_upper << 4) + md.u.md2.hwid,
			 bpf_ntohl(md.u.md2.timestamp));
#endif

	return TC_ACT_OK;
}

SEC("vxlan_set_tunnel")
int _vxlan_set_tunnel(struct __sk_buff *skb)
{
	int ret;
	struct bpf_tunnel_key key;
	struct vxlan_metadata md;

	__builtin_memset(&key, 0x0, sizeof(key));
	key.remote_ipv4 = 0xac100164; /* 172.16.1.100 */
	key.tunnel_id = 2;
	key.tunnel_tos = 0;
	key.tunnel_ttl = 64;

	ret = bpf_skb_set_tunnel_key(skb, &key, sizeof(key),
				     BPF_F_ZERO_CSUM_TX);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	md.gbp = 0x800FF; /* Set VXLAN Group Policy extension */
	ret = bpf_skb_set_tunnel_opt(skb, &md, sizeof(md));
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	return TC_ACT_OK;
}

SEC("vxlan_get_tunnel")
int _vxlan_get_tunnel(struct __sk_buff *skb)
{
	int ret;
	struct bpf_tunnel_key key;
	struct vxlan_metadata md;
	char fmt[] = "key %d remote ip 0x%x vxlan gbp 0x%x\n";

	ret = bpf_skb_get_tunnel_key(skb, &key, sizeof(key), 0);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	ret = bpf_skb_get_tunnel_opt(skb, &md, sizeof(md));
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	bpf_trace_printk(fmt, sizeof(fmt),
			key.tunnel_id, key.remote_ipv4, md.gbp);

	return TC_ACT_OK;
}

SEC("geneve_set_tunnel")
int _geneve_set_tunnel(struct __sk_buff *skb)
{
	int ret, ret2;
	struct bpf_tunnel_key key;
	struct geneve_opt gopt;

	__builtin_memset(&key, 0x0, sizeof(key));
	key.remote_ipv4 = 0xac100164; /* 172.16.1.100 */
	key.tunnel_id = 2;
	key.tunnel_tos = 0;
	key.tunnel_ttl = 64;

	__builtin_memset(&gopt, 0x0, sizeof(gopt));
	gopt.opt_class = bpf_htons(0x102); /* Open Virtual Networking (OVN) */
	gopt.type = 0x08;
	gopt.r1 = 0;
	gopt.r2 = 0;
	gopt.r3 = 0;
	gopt.length = 2; /* 4-byte multiple */
	*(int *) &gopt.opt_data = bpf_htonl(0xdeadbeef);

	ret = bpf_skb_set_tunnel_key(skb, &key, sizeof(key),
				     BPF_F_ZERO_CSUM_TX);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	ret = bpf_skb_set_tunnel_opt(skb, &gopt, sizeof(gopt));
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	return TC_ACT_OK;
}

SEC("geneve_get_tunnel")
int _geneve_get_tunnel(struct __sk_buff *skb)
{
	int ret;
	struct bpf_tunnel_key key;
	struct geneve_opt gopt;
	char fmt[] = "key %d remote ip 0x%x geneve class 0x%x\n";

	ret = bpf_skb_get_tunnel_key(skb, &key, sizeof(key), 0);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	ret = bpf_skb_get_tunnel_opt(skb, &gopt, sizeof(gopt));
	if (ret < 0)
		gopt.opt_class = 0;

	bpf_trace_printk(fmt, sizeof(fmt),
			key.tunnel_id, key.remote_ipv4, gopt.opt_class);
	return TC_ACT_OK;
}

// ipip_tunnel1 会封包，这里修改了 outer meta key.remote_ipv4, 来转发包 -> ipip_ns0 里的 ipip_veth0
SEC("ipip_set_tunnel")
int _ipip_set_tunnel(struct __sk_buff *skb)
{
	struct bpf_tunnel_key key = {};
	void *data = (void *)(long)skb->data;
	struct iphdr *iph = data;
	void *data_end = (void *)(long)skb->data_end;
	int ret;

	/* single length check */
	if (data + sizeof(*iph) > data_end) {
		ERROR(1);
		return TC_ACT_SHOT;
	}

	key.tunnel_ttl = 64;
	if (iph->protocol == IPPROTO_ICMP) {
		key.remote_ipv4 = 0xad100164; /* 173.16.1.100, ipip_veth0 网卡的 ip */
	}

	ret = bpf_skb_set_tunnel_key(skb, &key, sizeof(key), 0);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

	return TC_ACT_OK;
}

SEC("ipip_get_tunnel")
int _ipip_get_tunnel(struct __sk_buff *skb)
{
	int ret;
	struct bpf_tunnel_key key;
//	char fmt[] = "remote ip %s\n";
	char fmt[] = "remote ip 0x%x\n";

	ret = bpf_skb_get_tunnel_key(skb, &key, sizeof(key), 0);
	if (ret < 0) {
		ERROR(ret);
		return TC_ACT_SHOT;
	}

    // tail -n 100 /sys/kernel/debug/tracing/trace
//    char* ip_str = u32toIpStr(key.remote_ipv4);
	bpf_trace_printk(fmt, sizeof(fmt), key.remote_ipv4); // remote ip 0xad100164(=173.16.1.100)，这里是回包的 outer ip 地址
	return TC_ACT_OK;
}

SEC("xfrm_get_state")
int _xfrm_get_state(struct __sk_buff *skb)
{
	struct bpf_xfrm_state x;
	char fmt[] = "reqid %d spi 0x%x remote ip 0x%x\n";
	int ret;

	ret = bpf_skb_get_xfrm_state(skb, 0, &x, sizeof(x), 0);
	if (ret < 0)
		return TC_ACT_OK;

	bpf_trace_printk(fmt, sizeof(fmt), x.reqid, bpf_ntohl(x.spi),
			 bpf_ntohl(x.remote_ipv4));
	return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;
