

#ifndef __VIP_H
#define __VIP_H

#include <stddef.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/icmp.h>
#include <linux/icmpv6.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#include <loadbalancer/balancer_consts.h>
#include <loadbalancer/csum_helpers.h>


// vip's definition for lookup
struct vip_definition {
  union {
    __be32 vip;
    __be32 vipv6[4];
  };
  __u16 port;
  __u8 proto;
};

// result of vip's lookup
struct vip_meta {
  __u32 flags;
  __u32 vip_num;
};





// map, which contains all the vips for which we are doing load balancing，这个是关键 map!!!
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(key_size, sizeof(struct vip_definition));
	__uint(value_size, sizeof(struct vip_meta));
	__uint(max_entries, MAX_VIPS);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} vip_map SEC(".maps");




#endif