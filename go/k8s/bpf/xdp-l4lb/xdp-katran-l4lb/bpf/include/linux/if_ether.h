


#ifndef _UAPI_LINUX_IF_ETHER_H
#define _UAPI_LINUX_IF_ETHER_H

#include <linux/types.h>

// Ethernet header
// #include <linux/if_ether.h>
struct ethhdr {
  __u8 h_dest[6];
  __u8 h_source[6];
  __u16 h_proto;
} __attribute__((packed));


#endif