package payload

// Payload is the structure used when copying events from the main monitor.
type Payload struct {
    Data []byte
    CPU  int
    Lost uint64
    Type int
}
