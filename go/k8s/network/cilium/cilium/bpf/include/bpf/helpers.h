//
// Created by 刘祥 on 7/1/22.
//

#ifndef BPF_HELPERS_H
#define BPF_HELPERS_H


#include <linux/bpf.h>
#include "types_mapper.h"
#include "compiler.h"

#ifndef BPF_FUNC
# define BPF_FUNC(NAME, ...)						\
	(* NAME)(__VA_ARGS__) __maybe_unused = (void *)BPF_FUNC_##NAME
#endif




/* Sockops and SK_MSG helpers */
static int BPF_FUNC(sock_map_update, struct bpf_sock_ops *skops, void *map,
                    __u32 key,  __u64 flags);
        static int BPF_FUNC(sock_hash_update, struct bpf_sock_ops *skops, void *map,
                            void *key,  __u64 flags);
static int BPF_FUNC(msg_redirect_hash, struct sk_msg_md *md, void *map,
                    void *key, __u64 flags);

#endif //BPF_HELPERS_H
