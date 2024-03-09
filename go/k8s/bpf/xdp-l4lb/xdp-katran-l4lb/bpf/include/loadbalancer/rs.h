




#ifndef __RS_H
#define __RS_H

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




// where to send client's packet from LRU_MAP
struct real_pos_lru {
  __u32 pos;
  __u64 atime;
};

// where to send client's packet from lookup in ch ring.
struct real_definition {
  union {
    __be32 dst;
    __be32 dstv6[4];
  };
  __u8 flags;
};


// map which contains opaque real's id to real mapping, 这个是关键 map!!!
struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct real_definition));
	__uint(max_entries, MAX_REALS);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} reals SEC(".maps");






#endif