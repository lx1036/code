package server

import (
	"context"
	"github.com/eapache/channels"
	"github.com/osrg/gobgp/pkg/packet/bgp"
	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFSMHandlerEstablishHoldTimeExpired(t *testing.T) {
	m := NewMockConnection(t)
	p, h := makePeerAndHandler()

	// push mock connection
	p.fsm.conn = m
	p.fsm.h = h

	// set keepalive ticker
	p.fsm.pConf.Timers.State.NegotiatedHoldTime = 3

	msg := bgp.NewBGPKeepAliveMessage()
	header, _ := msg.Header.Serialize()
	body, _ := msg.Body.Serialize()

	pushPackets := func() {
		// first keepalive from peer
		m.setData(header)
		m.setData(body)
	}

	// set holdtime
	p.fsm.pConf.Timers.Config.HoldTime = 2
	p.fsm.pConf.Timers.State.NegotiatedHoldTime = 2

	// push keepalive msg
	go pushPackets()

	state, fsmStateReason := h.established(context.Background())
	time.Sleep(time.Second * 1)

	assert.Equal(t, bgp.BGP_FSM_IDLE, state)
	assert.Equal(t, fsmHoldTimerExpired, fsmStateReason.Type)

	m.mtx.Lock()
	lastMsg := m.sendBuf[len(m.sendBuf)-1]
	m.mtx.Unlock()
	sent, _ := bgp.ParseBGPMessage(lastMsg)
	assert.Equal(t, uint8(bgp.BGP_MSG_NOTIFICATION), sent.Header.Type)
	assert.Equal(t, uint8(bgp.BGP_ERROR_HOLD_TIMER_EXPIRED), sent.Body.(*bgp.BGPNotification).ErrorCode)
}

func makePeerAndHandler() (*peer, *fsmHandler) {
	p := &peer{
		fsm: newFSM(&config.Global{}, &config.Neighbor{}),
	}

	h := &fsmHandler{
		fsm:           p.fsm,
		stateReasonCh: make(chan fsmStateReason, 2),
		incoming:      channels.NewInfiniteChannel(),
		outgoing:      channels.NewInfiniteChannel(),
	}

	return p, h
}

type MockConnection struct {
	*testing.T
	net.Conn
	recvCh    chan chan byte
	sendBuf   [][]byte
	currentCh chan byte
	isClosed  bool
	wait      int
	mtx       sync.Mutex
}

func NewMockConnection(t *testing.T) *MockConnection {
	m := &MockConnection{
		T:        t,
		recvCh:   make(chan chan byte, 128),
		sendBuf:  make([][]byte, 0),
		isClosed: false,
	}
	return m
}

func (m *MockConnection) setData(data []byte) int {
	dataChan := make(chan byte, 4096)
	for _, b := range data {
		dataChan <- b
	}
	m.recvCh <- dataChan
	return len(dataChan)
}
