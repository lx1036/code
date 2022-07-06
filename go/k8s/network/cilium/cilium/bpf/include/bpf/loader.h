//
// Created by 刘祥 on 7/6/22.
//

#ifndef BPF_LOADER_H
#define BPF_LOADER_H


#include "types_mapper.h"




struct bpf_elf_map {
    __u32 type;
    __u32 size_key;
    __u32 size_value;
    __u32 max_elem;
    __u32 flags;
    __u32 id;
    __u32 pinning;
#ifdef SOCKMAP
    __u32 inner_id;
	__u32 inner_idx;
#endif
};





#endif //BPF_LOADER_H
