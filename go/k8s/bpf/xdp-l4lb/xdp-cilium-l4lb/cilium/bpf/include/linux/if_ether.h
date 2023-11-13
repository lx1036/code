


#ifndef _UAPI_LINUX_IF_ETHER_H
#define _UAPI_LINUX_IF_ETHER_H

#include <linux/types.h>



/*
 *	IEEE 802.3 Ethernet magic constants.  The frame sizes omit the preamble
 *	and FCS/CRC (frame check sequence).
 */

#define ETH_ALEN	6		/* Octets in one ethernet addr	 */
/* __ETH_HLEN is out of sync with the kernel's if_ether.h. In Cilium datapath
 * we use ETH_HLEN which can be loaded via static data, and for L2-less devs
 * it's 0. To avoid replacing every occurrence of ETH_HLEN in the datapath,
 * we prefixed the kernel's ETH_HLEN instead.
 */
// 二层头，14字节
#define __ETH_HLEN	14		/* Total octets in header.	 */
#define ETH_ZLEN	60		/* Min. octets in frame sans FCS */
#define ETH_DATA_LEN	1500		/* Max. octets in payload	 */
#define ETH_FRAME_LEN	1514		/* Max. octets in frame sans FCS */
#define ETH_FCS_LEN	4		/* Octets in the FCS		 */

#define ETH_P_802_3_MIN	0x0600		/* If the value in the ethernet type is less than this value
					 * then the frame is Ethernet II. Else it is 802.3 */


// Ethernet header
// #include <linux/if_ether.h>
struct ethhdr {
  __u8 h_dest[6];
  __u8 h_source[6];
  __u16 h_proto;
} __attribute__((packed));


#endif