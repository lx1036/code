

#ifndef __STATS_H
#define __STATS_H



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


// per vip statistics
struct lb_stats {
  __u64 v1;
  __u64 v2;
};


// map vip stats
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct lb_stats));
	__uint(max_entries, STATS_MAP_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} stats SEC(".maps");

// map with per real pps/bps statistic
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct lb_stats));
	__uint(max_entries, MAX_REALS);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} reals_stats SEC(".maps");




#endif