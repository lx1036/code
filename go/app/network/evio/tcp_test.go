package evio

// github.com/tidwall/evio

import (
	"github.com/tidwall/evio"
	"testing"
)

func TestName(test *testing.T) {
	var events evio.Events
	events.Data = func(connection evio.Conn, in []byte) (out []byte, action evio.Action) {
		out = in
		return
	}

	if err := evio.Serve(events, "tcp://localhost:5000"); err != nil {
		panic(err.Error())
	}
}
