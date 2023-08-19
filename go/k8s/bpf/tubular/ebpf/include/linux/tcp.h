

// /root/linux-5.10.142/include/uapi/linux/tcp.h



/* TCP socket options */
#define TCP_NODELAY		1	/* Turn off Nagle's algorithm. */
#define TCP_MAXSEG		2	/* Limit MSS */

#define TCP_KEEPIDLE		4	/* Start keeplives after this period */
#define TCP_KEEPINTVL		5	/* Interval between keepalives */
#define TCP_KEEPCNT		6	/* Number of keepalives before death */
#define TCP_SYNCNT		7	/* Number of SYN retransmits */
#define TCP_LINGER2		8	/* Life time of orphaned FIN-WAIT-2 state */

#define TCP_WINDOW_CLAMP	10	/* Bound advertised window */
#define TCP_INFO		11	/* Information about this connection. */


#define TCP_CONGESTION		13	/* Congestion control algorithm */

#define TCP_REPAIR		19	/* TCP sock is under repair right now */
#define TCP_REPAIR_QUEUE	20
#define TCP_QUEUE_SEQ		21

#define TCP_FASTOPEN		23	/* Enable FastOpen on listeners */

#define TCP_CC_INFO		26	/* Get Congestion Control (optional) info */
#define TCP_SAVE_SYN		27	/* Record SYN headers for new connections */
#define TCP_SAVED_SYN		28	/* Get SYN headers recorded for connection */

#define TCP_FASTOPEN_KEY	33	/* Set the key for Fast Open (cookie) */
#define TCP_FASTOPEN_NO_COOKIE	34	/* Enable TFO without a TFO cookie */
#define TCP_ZEROCOPY_RECEIVE	35

#define TCP_TX_DELAY		37	/* delay outgoing packets by XX usec */


// TCP header
struct tcphdr {
  __u16 source;
  __u16 dest;
  __u32 seq;
  __u32 ack_seq;
  union {
    struct {
      // Field order has been converted LittleEndiand -> BigEndian
      // in order to simplify flag checking (no need to ntohs())
      __u16 ns : 1,
      reserved : 3,
      doff : 4,
      fin : 1,
      syn : 1,
      rst : 1,
      psh : 1,
      ack : 1,
      urg : 1,
      ece : 1,
      cwr : 1;
    };
  };
  __u16 window;
  __u16 check;
  __u16 urg_ptr;
};
