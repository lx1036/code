


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

/*
 *	These are the defined Ethernet Protocol ID's.
 */
#define ETH_P_LOOP	0x0060		/* Ethernet Loopback packet	*/
#define ETH_P_PUP	0x0200		/* Xerox PUP packet		*/
#define ETH_P_PUPAT	0x0201		/* Xerox PUP Addr Trans packet	*/
#define ETH_P_TSN	0x22F0		/* TSN (IEEE 1722) packet	*/
#define ETH_P_IP	0x0800		/* Internet Protocol packet	*/
#define ETH_P_X25	0x0805		/* CCITT X.25			*/
#define ETH_P_ARP	0x0806		/* Address Resolution packet	*/
#define	ETH_P_BPQ	0x08FF		/* G8BPQ AX.25 Ethernet Packet	[ NOT AN OFFICIALLY REGISTERED ID ] */
#define ETH_P_IEEEPUP	0x0a00		/* Xerox IEEE802.3 PUP packet */
#define ETH_P_IEEEPUPAT	0x0a01		/* Xerox IEEE802.3 PUP Addr Trans packet */
#define ETH_P_BATMAN	0x4305		/* B.A.T.M.A.N.-Advanced packet [ NOT AN OFFICIALLY REGISTERED ID ] */
#define ETH_P_DEC       0x6000          /* DEC Assigned proto           */
#define ETH_P_DNA_DL    0x6001          /* DEC DNA Dump/Load            */
#define ETH_P_DNA_RC    0x6002          /* DEC DNA Remote Console       */
#define ETH_P_DNA_RT    0x6003          /* DEC DNA Routing              */
#define ETH_P_LAT       0x6004          /* DEC LAT                      */
#define ETH_P_DIAG      0x6005          /* DEC Diagnostics              */
#define ETH_P_CUST      0x6006          /* DEC Customer use             */
#define ETH_P_SCA       0x6007          /* DEC Systems Comms Arch       */
#define ETH_P_TEB	0x6558		/* Trans Ether Bridging		*/
#define ETH_P_RARP      0x8035		/* Reverse Addr Res packet	*/
#define ETH_P_ATALK	0x809B		/* Appletalk DDP		*/
#define ETH_P_AARP	0x80F3		/* Appletalk AARP		*/
#define ETH_P_8021Q	0x8100          /* 802.1Q VLAN Extended Header  */
#define ETH_P_IPX	0x8137		/* IPX over DIX			*/
#define ETH_P_IPV6	0x86DD		/* IPv6 over bluebook		*/
#define ETH_P_PAUSE	0x8808		/* IEEE Pause frames. See 802.3 31B */
#define ETH_P_SLOW	0x8809		/* Slow Protocol. See 802.3ad 43B */
#define ETH_P_WCCP	0x883E		/* Web-cache coordination protocol defined in draft-wilson-wrec-wccp-v2-00.txt */
#define ETH_P_MPLS_UC	0x8847		/* MPLS Unicast traffic		*/
#define ETH_P_MPLS_MC	0x8848		/* MPLS Multicast traffic	*/
#define ETH_P_ATMMPOA	0x884c		/* MultiProtocol Over ATM	*/
#define ETH_P_PPP_DISC	0x8863		/* PPPoE discovery messages     */
#define ETH_P_PPP_SES	0x8864		/* PPPoE session messages	*/
#define ETH_P_LINK_CTL	0x886c		/* HPNA, wlan link local tunnel */
#define ETH_P_ATMFATE	0x8884		/* Frame-based ATM Transport * over Ethernet */
#define ETH_P_PAE	0x888E		/* Port Access Entity (IEEE 802.1X) */
#define ETH_P_AOE	0x88A2		/* ATA over Ethernet		*/
#define ETH_P_8021AD	0x88A8          /* 802.1ad Service VLAN		*/
#define ETH_P_802_EX1	0x88B5		/* 802.1 Local Experimental 1.  */
#define ETH_P_TIPC	0x88CA		/* TIPC 			*/
#define ETH_P_8021AH	0x88E7          /* 802.1ah Backbone Service Tag */
#define ETH_P_MVRP	0x88F5          /* 802.1Q MVRP                  */
#define ETH_P_1588	0x88F7		/* IEEE 1588 Timesync */
#define ETH_P_PRP	0x88FB		/* IEC 62439-3 PRP/HSRv0	*/
#define ETH_P_FCOE	0x8906		/* Fibre Channel over Ethernet  */
#define ETH_P_TDLS	0x890D          /* TDLS */
#define ETH_P_FIP	0x8914		/* FCoE Initialization Protocol */
#define ETH_P_80221	0x8917		/* IEEE 802.21 Media Independent Handover Protocol */
#define ETH_P_LOOPBACK	0x9000		/* Ethernet loopback packet, per IEEE 802.3 */
#define ETH_P_QINQ1	0x9100		/* deprecated QinQ VLAN [ NOT AN OFFICIALLY REGISTERED ID ] */
#define ETH_P_QINQ2	0x9200		/* deprecated QinQ VLAN [ NOT AN OFFICIALLY REGISTERED ID ] */
#define ETH_P_QINQ3	0x9300		/* deprecated QinQ VLAN [ NOT AN OFFICIALLY REGISTERED ID ] */
#define ETH_P_EDSA	0xDADA		/* Ethertype DSA [ NOT AN OFFICIALLY REGISTERED ID ] */
#define ETH_P_AF_IUCV   0xFBFB		/* IBM af_iucv [ NOT AN OFFICIALLY REGISTERED ID ] */
#define ETH_P_802_3_MIN	0x0600		/* If the value in the ethernet type is less than this value then the frame is Ethernet II. Else it is 802.3 */


// Ethernet header
// #include <linux/if_ether.h>
struct ethhdr {
  __u8 h_dest[6];
  __u8 h_source[6];
  __u16 h_proto; // 二层头里包含协议字段
} __attribute__((packed));


#endif