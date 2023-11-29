



#ifndef _UAPI_IPV6_H
#define _UAPI_IPV6_H

#include <linux/types.h>
#include <linux/in6.h>


#define GET_PREFIX(PREFIX)  bpf_htonl(PREFIX <= 0 ? 0 : PREFIX < 32 ? ((1<<PREFIX) - 1) << (32-PREFIX)	: 0xFFFFFFFF)


/*
 *	IPv6 fixed header
 *
 *	BEWARE, it is incorrect. The first 4 bits of flow_lbl
 *	are glued to priority now, forming "class".
 */

struct ipv6hdr {
// #if defined(__LITTLE_ENDIAN_BITFIELD)
	// __u8			priority:4,
				// version:4;
// #elif defined(__BIG_ENDIAN_BITFIELD)
	
// #endif
	__u8			version:4,
				priority:4;
	__u8			flow_lbl[3];

	__be16			payload_len;
	__u8			nexthdr;
	__u8			hop_limit;

	struct	in6_addr	saddr;
	struct	in6_addr	daddr;
};





#endif /* _UAPI_IPV6_H */
