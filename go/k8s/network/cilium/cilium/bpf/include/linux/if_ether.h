//
// Created by 刘祥 on 7/1/22.
//

#ifndef _LINUX_IF_ETHER_H
#define _LINUX_IF_ETHER_H


/*
 *	IEEE 802.3 Ethernet magic constants.  The frame sizes omit the preamble
 *	and FCS/CRC (frame check sequence).
 */

#define ETH_ALEN 6		/* Octets in one ethernet addr	 */
#define ETH_HLEN 14		/* Total octets in header.	 */
#define ETH_P_IP	0x0800		/* Internet Protocol packet	*/
#define ETH_P_X25	0x0805		/* CCITT X.25			*/
#define ETH_P_ARP	0x0806		/* Address Resolution packet	*/

#define ETH_P_802_3_MIN	0x0600		/* If the value in the ethernet type is less than this value
					 * then the frame is Ethernet II. Else it is 802.3 */

/*
 *	This is an Ethernet frame header.
 */

struct ethhdr {
    unsigned char	h_dest[ETH_ALEN];	/* destination eth addr	*/
    unsigned char	h_source[ETH_ALEN];	/* source ether addr	*/
    __be16		h_proto;		/* packet type ID field	*/
} __attribute__((packed));



#endif //_LINUX_IF_ETHER_H
