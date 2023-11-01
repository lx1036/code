



#ifndef __BPF_CTX_COMMON_H_
#define __BPF_CTX_COMMON_H_

#include <linux/types.h>



#define __ctx_skb		1
#define __ctx_xdp		2



























/////////////////////补充定义///////////////////////////

/*
 * Helper macro to place programs, maps, license in
 * different sections in elf_bpf file. Section names
 * are interpreted by elf_bpf loader
 */
#define SEC(NAME) __attribute__((section(NAME), used))







/////////////////////补充定义///////////////////////////


#endif /* __BPF_CTX_COMMON_H_ */