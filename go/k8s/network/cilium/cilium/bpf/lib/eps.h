//
// Created by 刘祥 on 7/5/22.
//

#ifndef BPF_EPS_H
#define BPF_EPS_H

#include "bpf/compiler.h"
#include "bpf/types_mapper.h"
#include "bpf/helpers.h"
#include "maps.h"
#include "common.h"



// EP_POLICY_MAP 会在 node_config.h 中定义为 cilium_ep_to_policy
// 查看 go/k8s/network/cilium/cilium/pkg/bpf/maps/endpointpolicymap
// 根据 ip 查看属于的 endpoint
static __always_inline void *
lookup_ip4_endpoint_policy_map(__u32 ip)
{
    struct endpoint_key key = {};

    key.ip4 = ip;
    key.family = ENDPOINT_KEY_IPV4;

    return map_lookup_elem(&EP_POLICY_MAP, &key);
}


static __always_inline __maybe_unused struct remote_endpoint_info *
ipcache_lookup4(struct bpf_elf_map *map, __be32 addr, __u32 prefix)
{

}


#define lookup_ip4_remote_endpoint(addr) \
	ipcache_lookup4(&IPCACHE_MAP, addr, V4_CACHE_KEY_LEN)




#endif //BPF_EPS_H
