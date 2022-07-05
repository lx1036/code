//
// Created by 刘祥 on 7/1/22.
//

#ifndef __LIB_ENDIAN_H_
#define __LIB_ENDIAN_H_

#include <linux/byteorder/little_endian.h>


# define __bpf_ntohs(x)		__builtin_bswap16(x)
# define __bpf_htons(x)		__builtin_bswap16(x)
# define __bpf_ntohl(x)		__builtin_bswap32(x)
# define __bpf_htonl(x)		__builtin_bswap32(x)

#define bpf_ntohs(x) (__builtin_constant_p(x) ?	__constant_ntohs(x) : __bpf_ntohs(x))
#define bpf_htons(x) (__builtin_constant_p(x) ? __constant_htons(x) : __bpf_htons(x))
#define bpf_htonl(x) (__builtin_constant_p(x) ? __constant_htonl(x) : __bpf_htonl(x))
#define bpf_ntohl(x)				\
	(__builtin_constant_p(x) ?		\
	 __constant_ntohl(x) : __bpf_ntohl(x))


#endif //__LIB_ENDIAN_H_
