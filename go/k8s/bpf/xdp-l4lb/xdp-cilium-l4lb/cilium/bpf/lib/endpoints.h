

#ifndef XDP_CILIUM_L4LB_EPS_H
#define XDP_CILIUM_L4LB_EPS_H


#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/bpf.h>

#include "maps.h"


#define V4_CACHE_KEY_LEN (sizeof(__u32)*8)

/* IPCACHE_STATIC_PREFIX gets sizeof non-IP, non-prefix part of ipcache_key */
#define IPCACHE_STATIC_PREFIX   (8 * (sizeof(struct ipcache_key) - sizeof(struct bpf_lpm_trie_key) - sizeof(union v6addr)))
#define IPCACHE_PREFIX_LEN(PREFIX) (IPCACHE_STATIC_PREFIX + (PREFIX))

#define lookup_ip4_remote_endpoint(addr) ipcache_lookup4(&IPCACHE_MAP, addr, V4_CACHE_KEY_LEN)

static __always_inline __maybe_unused
struct endpoint_info * __lookup_ip4_endpoint(__u32 ip) {
    struct endpoint_key key = {};

    key.ip4 = ip;
    key.family = ENDPOINT_KEY_IPV4;

    return bpf_map_lookup_elem(&cilium_lxc, &key);
}

static __always_inline __maybe_unused
struct endpoint_info * lookup_ip4_endpoint(const struct iphdr *ip4) {
    return __lookup_ip4_endpoint(ip4->daddr);
}

static __always_inline __maybe_unused
struct remote_endpoint_info * ipcache_lookup4(__be32 addr) {
    __u32 prefix = V4_CACHE_KEY_LEN;
    struct ipcache_key key = {
            .lpm_key = { IPCACHE_PREFIX_LEN(prefix), {} },
            .family = ENDPOINT_KEY_IPV4,
            .ip4 = addr,
    };
    key.ip4 &= GET_PREFIX(prefix);
    return bpf_map_lookup_elem(&cilium_ipcache, &key);
}







#endif //XDP_CILIUM_L4LB_EPS_H
