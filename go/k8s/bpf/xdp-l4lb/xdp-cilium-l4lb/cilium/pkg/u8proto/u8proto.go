package u8proto

type U8proto uint8

const (
	// ANY represents all protocols.
	ANY    U8proto = 0
	ICMP   U8proto = 1
	TCP    U8proto = 6
	UDP    U8proto = 17
	ICMPv6 U8proto = 58
)
