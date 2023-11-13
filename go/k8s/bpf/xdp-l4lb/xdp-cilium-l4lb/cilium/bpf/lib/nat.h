



/* Simple NAT engine in BPF. */
#ifndef __LIB_NAT__
#define __LIB_NAT__


#include <linux/icmp.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/ip.h>
// #include <linux/icmpv6.h>
// #include <linux/ipv6.h>

#include "common.h"
#include "drop.h"
#include "signal.h"
#include "conntrack.h"
#include "conntrack_map.h"
// #include "icmp6.h"



struct ipv4_nat_target {
	__be32 addr;
	const __u16 min_port; /* host endianness */
	const __u16 max_port; /* host endianness */
	bool src_from_world;
};



#endif /* __LIB_NAT__ */