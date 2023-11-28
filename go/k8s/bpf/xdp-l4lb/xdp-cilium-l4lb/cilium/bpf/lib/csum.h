

#ifndef __LIB_CSUM_H_
#define __LIB_CSUM_H_

#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/icmpv6.h>
#include <linux/in.h>
#include <linux/in6.h>



#define TCP_CSUM_OFF (offsetof(struct tcphdr, check))
#define UDP_CSUM_OFF (offsetof(struct udphdr, check))

struct csum_offset {
    __u16 offset;
    __u16 flags;
};

/**
 * Determins the L4 checksum field offset and required flags
 * @arg nexthdr	L3 nextheader field
 * @arg off	Pointer to uninitialied struct csum_offset struct
 *
 * Sets off.offset to offset from start of L4 header to L4 checksum field
 * and off.flags to the required flags, namely BPF_F_MARK_MANGLED_0 for UDP.
 * For unknown L4 protocols or L4 protocols which do not have a checksum
 * field, off is initialied to 0.
 */
static __always_inline void csum_l4_offset_and_flags(__u8 nexthdr, struct csum_offset *off)
{
    switch(nexthdr) {
        case IPPROTO_TCP:
            off->offset = TCP_CSUM_OFF;
            break;

        case IPPROTO_UDP:
            off->offset = UDP_CSUM_OFF;
            off->flags = BPF_F_MARK_MANGLED_0;
            break;

        case IPPROTO_ICMPV6:
            off->offset = offsetof(struct icmp6hdr, icmp6_cksum);
            break;

        case IPPROTO_ICMP:
        default:
            break;
    }
}




#endif //__LIB_CSUM_H_
