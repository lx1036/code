package server

import (
	"context"
	"fmt"
	"github.com/eapache/channels"
	"github.com/osrg/gobgp/pkg/packet/bgp"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/table"
	"net"
	"sync"
	"time"
)

const (
	minConnectRetryInterval = 5
)

type fsmStateReasonType uint8

const (
	fsmDying fsmStateReasonType = iota
	fsmAdminDown
	fsmReadFailed
	fsmWriteFailed
	fsmNotificationSent
	fsmNotificationRecv
	fsmHoldTimerExpired
	fsmIdleTimerExpired
	fsmRestartTimerExpired
	fsmGracefulRestart
	fsmInvalidMsg
	fsmNewConnection
	fsmOpenMsgReceived
	fsmOpenMsgNegotiated
	fsmHardReset
	fsmDeConfigured
)

const (
	holdtimeOpensent = 240
	holdtimeIdle     = 5
)

type fsmStateReason struct {
	Type            fsmStateReasonType
	BGPNotification *bgp.BGPMessage
	Data            []byte
}

func newfsmStateReason(typ fsmStateReasonType, notif *bgp.BGPMessage, data []byte) *fsmStateReason {
	return &fsmStateReason{
		Type:            typ,
		BGPNotification: notif,
		Data:            data,
	}
}

func (r fsmStateReason) String() string {
	switch r.Type {
	case fsmDying:
		return "dying"
	case fsmAdminDown:
		return "admin-down"
	case fsmReadFailed:
		return "read-failed"
	case fsmWriteFailed:
		return "write-failed"
	case fsmNotificationSent:
		body := r.BGPNotification.Body.(*bgp.BGPNotification)
		return fmt.Sprintf("notification-sent %s", bgp.NewNotificationErrorCode(body.ErrorCode, body.ErrorSubcode).String())
	case fsmNotificationRecv:
		body := r.BGPNotification.Body.(*bgp.BGPNotification)
		return fmt.Sprintf("notification-received %s", bgp.NewNotificationErrorCode(body.ErrorCode, body.ErrorSubcode).String())
	case fsmHoldTimerExpired:
		return "hold-timer-expired"
	case fsmIdleTimerExpired:
		return "idle-hold-timer-expired"
	case fsmRestartTimerExpired:
		return "restart-timer-expired"
	case fsmGracefulRestart:
		return "graceful-restart"
	case fsmInvalidMsg:
		return "invalid-msg"
	case fsmNewConnection:
		return "new-connection"
	case fsmOpenMsgReceived:
		return "open-msg-received"
	case fsmOpenMsgNegotiated:
		return "open-msg-negotiated"
	case fsmHardReset:
		return "hard-reset"
	default:
		return "unknown"
	}
}

type adminState int

const (
	adminStateUp adminState = iota
	adminStateDown
	adminStatePfxCt
)

func (s adminState) String() string {
	switch s {
	case adminStateUp:
		return "adminStateUp"
	case adminStateDown:
		return "adminStateDown"
	case adminStatePfxCt:
		return "adminStatePfxCt"
	default:
		return "Unknown"
	}
}

type adminStateOperation struct {
	State         adminState
	Communication []byte
}

type fsm struct {
	lock sync.RWMutex

	gConf                *config.Global
	pConf                *config.Neighbor
	state                bgp.FSMState
	outgoingCh           *channels.InfiniteChannel
	incomingCh           *channels.InfiniteChannel
	reason               *fsmStateReason
	conn                 net.Conn
	connCh               chan net.Conn
	idleHoldTime         float64
	opensentHoldTime     float64
	adminState           adminState
	adminStateCh         chan adminStateOperation
	h                    *fsmHandler
	rfMap                map[bgp.RouteFamily]bgp.BGPAddPathMode
	capMap               map[bgp.BGPCapabilityCode][]bgp.ParameterCapabilityInterface
	recvOpen             *bgp.BGPMessage
	peerInfo             *table.PeerInfo
	gracefulRestartTimer *time.Timer
	twoByteAsTrans       bool
	marshallingOptions   *bgp.MarshallingOption
	notification         chan *bgp.BGPMessage
}

func newFSM(gConf *config.Global, pConf *config.Neighbor) *fsm {
	adminState := adminStateUp
	if pConf.Config.AdminDown {
		adminState = adminStateDown
	}
	pConf.State.SessionState = config.IntToSessionStateMap[int(bgp.BGP_FSM_IDLE)]
	pConf.Timers.State.Downtime = time.Now().Unix()
	fsm := &fsm{
		gConf:                gConf,
		pConf:                pConf,
		state:                bgp.BGP_FSM_IDLE,
		outgoingCh:           channels.NewInfiniteChannel(),
		incomingCh:           channels.NewInfiniteChannel(),
		connCh:               make(chan net.Conn, 1),
		opensentHoldTime:     float64(holdtimeOpensent),
		adminState:           adminState,
		adminStateCh:         make(chan adminStateOperation, 1),
		rfMap:                make(map[bgp.RouteFamily]bgp.BGPAddPathMode),
		capMap:               make(map[bgp.BGPCapabilityCode][]bgp.ParameterCapabilityInterface),
		peerInfo:             table.NewPeerInfo(gConf, pConf),
		gracefulRestartTimer: time.NewTimer(time.Hour),
		notification:         make(chan *bgp.BGPMessage, 1),
	}
	fsm.gracefulRestartTimer.Stop()
	return fsm
}

func (f *fsm) established() {

}

type fsmMsgType int

const (
	_ fsmMsgType = iota
	fsmMsgStateChange
	fsmMsgBGPMessage
	fsmMsgRouteRefresh
)

type fsmMsg struct {
	MsgType     fsmMsgType
	fsm         *fsm
	MsgSrc      string
	MsgData     interface{}
	StateReason *fsmStateReason
	PathList    []*table.Path
	timestamp   time.Time
	payload     []byte
}

type fsmHandler struct {
	fsm *fsm

	//incoming chan *fsmMsg
	incoming *channels.InfiniteChannel
	outgoing *channels.InfiniteChannel

	conn          net.Conn
	msgCh         *channels.InfiniteChannel
	stateReasonCh chan fsmStateReason
	//outgoing         *channels.InfiniteChannel
	holdTimerResetCh chan bool
	sentNotification *bgp.BGPMessage
	ctx              context.Context
	ctxCancel        context.CancelFunc
	wg               *sync.WaitGroup
}

func newFSMHandler(fsm *fsm, outgoing *channels.InfiniteChannel) *fsmHandler {
	ctx, cancel := context.WithCancel(context.Background())
	h := &fsmHandler{
		fsm:              fsm,
		stateReasonCh:    make(chan fsmStateReason, 2),
		incoming:         fsm.incomingCh,
		outgoing:         outgoing,
		holdTimerResetCh: make(chan bool, 2),
		wg:               &sync.WaitGroup{},
		ctx:              ctx,
		ctxCancel:        cancel,
	}
	h.wg.Add(1)

	go h.run(ctx)

	return h
}

func (h *fsmHandler) run(ctx context.Context) error {
	defer h.wg.Done()

	fsm := h.fsm
	fsm.lock.RLock()
	oldState := fsm.state
	fsm.lock.RUnlock()

	var reason *fsmStateReason
	nextState := bgp.FSMState(-1)
	fsm.lock.RLock()
	fsmState := fsm.state
	fsm.lock.RUnlock()

	switch fsmState {
	case bgp.BGP_FSM_IDLE:
		//nextState, reason = h.idle(ctx)
		// case bgp.BGP_FSM_CONNECT:
		// 	nextState = h.connect()
	case bgp.BGP_FSM_ACTIVE:
		//nextState, reason = h.active(ctx)
	case bgp.BGP_FSM_OPENSENT:
		//nextState, reason = h.opensent(ctx)
	case bgp.BGP_FSM_OPENCONFIRM:
		//nextState, reason = h.openconfirm(ctx)
	case bgp.BGP_FSM_ESTABLISHED:
		nextState, reason = h.established(ctx)
	}

	fsm.lock.RLock()
	fsm.reason = reason
	if nextState == bgp.BGP_FSM_ESTABLISHED && oldState == bgp.BGP_FSM_OPENCONFIRM {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   fsm.pConf.State.NeighborAddress,
			"State": fsm.state.String(),
		}).Info("Peer Up")
	}
	if oldState == bgp.BGP_FSM_ESTABLISHED {
		// The main goroutine sent the notification due to
		// deconfiguration or something.
		reason := fsm.reason
		if fsm.h.sentNotification != nil {
			reason.Type = fsmNotificationSent
			reason.BGPNotification = fsm.h.sentNotification
		}
		log.WithFields(log.Fields{
			"Topic":  "Peer",
			"Key":    fsm.pConf.State.NeighborAddress,
			"State":  fsm.state.String(),
			"Reason": reason.String(),
		}).Info("Peer Down")
	}
	fsm.lock.RUnlock()

	fsm.lock.RLock()
	h.incoming.In() <- &fsmMsg{
		fsm:         fsm,
		MsgType:     fsmMsgStateChange,
		MsgSrc:      fsm.pConf.State.NeighborAddress,
		MsgData:     nextState,
		StateReason: reason,
	}
	fsm.lock.RUnlock()

	return nil
}

func (h *fsmHandler) established(ctx context.Context) (bgp.FSMState, *fsmStateReason) {
	var wg sync.WaitGroup
	fsm := h.fsm
	fsm.lock.Lock()
	h.conn = fsm.conn
	fsm.lock.Unlock()

	defer wg.Wait()
	wg.Add(2)

	go h.sendMessageloop(ctx, &wg)
	h.msgCh = h.incoming
	go h.recvMessageloop(ctx, &wg)

	var holdTimer *time.Timer
	if fsm.pConf.Timers.State.NegotiatedHoldTime == 0 {
		holdTimer = &time.Timer{}
	} else {
		fsm.lock.RLock()
		holdTimer = time.NewTimer(time.Second * time.Duration(fsm.pConf.Timers.State.NegotiatedHoldTime))
		fsm.lock.RUnlock()
	}

	fsm.gracefulRestartTimer.Stop()

	for {
		select {}
	}
}

func (h *fsmHandler) sendMessageloop(ctx context.Context, wg *sync.WaitGroup) error {
	sendToStateReasonCh := func(typ fsmStateReasonType, notif *bgp.BGPMessage) {
		// probably doesn't happen but be cautious
		select {
		case h.stateReasonCh <- *newfsmStateReason(typ, notif, nil):
		default:
		}
	}

	defer wg.Done()
	conn := h.conn
	fsm := h.fsm
	ticker := keepaliveTicker(fsm)

	for {
		select {
		case <-ctx.Done():
			return nil
		case o := <-h.outgoing.Out():
			switch m := o.(type) {
			case *fsmOutgoingMsg:
				h.fsm.lock.RLock()
				options := h.fsm.marshallingOptions
				h.fsm.lock.RUnlock()
				for _, msg := range table.CreateUpdateMsgFromPaths(m.Paths, options) {
					if err := send(msg); err != nil {
						return nil
					}
				}
				if m.Notification != nil {
					if m.StayIdle {
						// current user is only prefix-limit
						// fix me if this is not the case
						h.changeadminState(adminStatePfxCt)
					}
					if err := send(m.Notification); err != nil {
						return nil
					}
				}
			default:
				return nil
			}
		case <-ticker.C:
			if err := send(bgp.NewBGPKeepAliveMessage()); err != nil {
				return nil
			}
		}
	}
}

func (h *fsmHandler) recvMessageloop(ctx context.Context, wg *sync.WaitGroup) error {

}

func keepaliveTicker(fsm *fsm) *time.Ticker {
	fsm.lock.RLock()
	defer fsm.lock.RUnlock()

	negotiatedTime := fsm.pConf.Timers.State.NegotiatedHoldTime
	if negotiatedTime == 0 {
		return &time.Ticker{}
	}
	sec := time.Second * time.Duration(fsm.pConf.Timers.State.KeepaliveInterval)
	if sec == 0 {
		sec = time.Second
	}
	return time.NewTicker(sec)
}
