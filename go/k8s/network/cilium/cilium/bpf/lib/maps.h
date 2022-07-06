//
// Created by 刘祥 on 7/1/22.
//

#ifndef __LIB_MAPS_H_
#define __LIB_MAPS_H_


#include "bpf/compiler.h"
#include "bpf/loader.h"
#include "bpf/section.h"
#include <linux/bpf.h>


/* Map to link endpoint id to per endpoint cilium_policy map */
struct bpf_elf_map __section_maps EP_POLICY_MAP = {
        .type = BPF_MAP_TYPE_HASH_OF_MAPS,
        .size_key = sizeof(struct endpoint_key),
        .size_value = sizeof(int),
        .pinning = PIN_GLOBAL_NS,
        .max_elem = ENDPOINTS_MAP_SIZE,
};




#endif //__LIB_MAPS_H_
