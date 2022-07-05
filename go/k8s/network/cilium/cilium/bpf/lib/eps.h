//
// Created by 刘祥 on 7/5/22.
//

#ifndef BPF_EPS_H
#define BPF_EPS_H




static __always_inline __maybe_unused struct remote_endpoint_info *
ipcache_lookup4(struct bpf_elf_map *map, __be32 addr, __u32 prefix)
{

}


#define lookup_ip4_remote_endpoint(addr) \
	ipcache_lookup4(&IPCACHE_MAP, addr, V4_CACHE_KEY_LEN)




#endif //BPF_EPS_H
