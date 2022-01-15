package raft

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/hashicorp/go-msgpack/codec"
)

func init() {
	// Ensure we use a high-entropy seed for the pseudo-random generator
	rand.Seed(newSeed())
}

// returns an int64 from a crypto random source
// can be used to seed a source for a math/rand.
func newSeed() int64 {
	r, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}
	return r.Int64()
}

// randomTimeout returns a value that is between the minVal and 2x minVal.
func randomTimeout(minVal time.Duration) <-chan time.Time {
	if minVal == 0 {
		return nil
	}
	extra := time.Duration(rand.Int63()) % minVal
	return time.After(minVal + extra)
}

// Decode reverses the encode operation on a byte slice input.
func decodeMsgPack(buf []byte, out interface{}) error {
	r := bytes.NewBuffer(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}

// Encode writes an encoded object to a new bytes buffer.
func encodeMsgPack(in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}

// max returns the maximum.
func max(a, b uint64) uint64 {
	if a >= b {
		return a
	}
	return b
}

// min returns the minimum.
func min(a, b uint64) uint64 {
	if a <= b {
		return a
	}
	return b
}
