//
// Created by 刘祥 on 7/1/22.
//

#ifndef __BPF_TYPES_MAPPER__
#define __BPF_TYPES_MAPPER__


typedef __signed__ char __s8;
typedef unsigned char __u8;

typedef __signed__ short __s16;
typedef unsigned short __u16;

typedef __u16 __le16;
typedef __u16 __be16;

typedef __signed__ int __s32;
typedef unsigned int __u32;

typedef __u32 __le32;
typedef __u32 __be32;


typedef __u16 __sum16;
typedef __u32 __wsum;

#endif //__BPF_TYPES_MAPPER__
