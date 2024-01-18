

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


char _license[] SEC("license") = "GPL";
int _version SEC("version") = 1;
