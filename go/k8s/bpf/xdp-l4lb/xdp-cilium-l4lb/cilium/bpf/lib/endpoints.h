



#ifndef __LIB_EPS_H_
#define __LIB_EPS_H_


#include <linux/ip.h>
#include <linux/ipv6.h>

#include "maps.h"



static __always_inline __maybe_unused 
struct endpoint_info * __lookup_ip4_endpoint(__u32 ip) {
	struct endpoint_key key = {};

	key.ip4 = ip;
	key.family = ENDPOINT_KEY_IPV4;

	return map_lookup_elem(&ENDPOINTS_MAP, &key);
}

static __always_inline __maybe_unused 
struct endpoint_info * lookup_ip4_endpoint(const struct iphdr *ip4) {
	return __lookup_ip4_endpoint(ip4->daddr);
}


static __always_inline __maybe_unused struct remote_endpoint_info *
ipcache_lookup4(struct bpf_elf_map *map, __be32 addr, __u32 prefix)
{
	struct ipcache_key key = {
		.lpm_key = { IPCACHE_PREFIX_LEN(prefix), {} },
		.family = ENDPOINT_KEY_IPV4,
		.ip4 = addr,
	};
	key.ip4 &= GET_PREFIX(prefix);
	return map_lookup_elem(map, &key);
}


#endif /* __LIB_EPS_H_ */