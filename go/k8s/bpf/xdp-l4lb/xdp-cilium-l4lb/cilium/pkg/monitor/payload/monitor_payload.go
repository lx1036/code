package payload

import "encoding/gob"

// Below constants are based on the ones from <linux/perf_event.h>.
// /root/linux-5.10.142/tools/include/uapi/linux/perf_event.h
const (
	// EventSample is equivalent to PERF_RECORD_SAMPLE
	EventSample = 9
	// RecordLost is equivalent to PERF_RECORD_LOST
	RecordLost = 2
)

// Payload is the structure used when copying events from the main monitor.
type Payload struct {
	Data []byte
	CPU  int
	Lost uint64
	Type int
}

// EncodeBinary writes the payload into its binary representation.
func (pl *Payload) EncodeBinary(enc *gob.Encoder) error {
	return enc.Encode(pl)
}

// DecodeBinary reads the payload from its binary representation.
func (pl *Payload) DecodeBinary(dec *gob.Decoder) error {
	return dec.Decode(pl)
}
