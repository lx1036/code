//
// Created by 刘祥 on 7/1/22.
//

#ifndef BPF_LITTLE_ENDIAN_H
#define BPF_LITTLE_ENDIAN_H

#include <linux/swab.h>

// https://github.com/torvalds/linux/blob/v5.19-rc4/include/uapi/linux/byteorder/little_endian.h#L16-L19

#define __constant_htonl(x) ((__be32)___constant_swab32((x)))
#define __constant_ntohl(x) ___constant_swab32((__be32)(x))
#define __constant_htons(x) ((__be16)___constant_swab16((x)))
#define __constant_ntohs(x) ___constant_swab16((__be16)(x))

#endif //BPF_LITTLE_ENDIAN_H
