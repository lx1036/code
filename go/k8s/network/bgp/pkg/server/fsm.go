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

func (fsm *fsm) bgpMessageStateUpdate(MessageType uint8, isIn bool) {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()
	state := &fsm.pConf.State.Messages
	timer := &fsm.pConf.Timers
	if isIn {
		state.Received.Total++
	} else {
		state.Sent.Total++
	}
	switch MessageType {
	case bgp.BGP_MSG_OPEN:
		if isIn {
			state.Received.Open++
		} else {
			state.Sent.Open++
		}
	case bgp.BGP_MSG_UPDATE:
		if isIn {
			state.Received.Update++
			timer.State.UpdateRecvTime = time.Now().Unix()
		} else {
			state.Sent.Update++
		}
	case bgp.BGP_MSG_NOTIFICATION:
		if isIn {
			state.Received.Notification++
		} else {
			state.Sent.Notification++
		}
	case bgp.BGP_MSG_KEEPALIVE:
		if isIn {
			state.Received.Keepalive++
		} else {
			state.Sent.Keepalive++
		}
	case bgp.BGP_MSG_ROUTE_REFRESH:
		if isIn {
			state.Received.Refresh++
		} else {
			state.Sent.Refresh++
		}
	default:
		if isIn {
			state.Received.Discarded++
		} else {
			state.Sent.Discarded++
		}
	}
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
		select {
		case <-ctx.Done():
			select {
			case m := <-fsm.notification:
				b, _ := m.Serialize(h.fsm.marshallingOptions)
				h.conn.Write(b)
			default:
				// nothing to do
			}
			h.conn.Close()
			return -1, newfsmStateReason(fsmDying, nil, nil)
		case conn, ok := <-fsm.connCh:
			if !ok {
				break
			}
			conn.Close()
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
			}).Warn("Closed an accepted connection")
			fsm.lock.RUnlock()
		case err := <-h.stateReasonCh:
			h.conn.Close()
			// if recv goroutine hit an error and sent to
			// stateReasonCh, then tx goroutine might take
			// long until it exits because it waits for
			// ctx.Done() or keepalive timer. So let kill
			// it now.
			h.outgoing.In() <- err
			fsm.lock.RLock()
			if s := fsm.pConf.GracefulRestart.State; s.Enabled {
				if (s.NotificationEnabled && err.Type == fsmNotificationRecv) ||
					(err.Type == fsmNotificationSent &&
						err.BGPNotification.Body.(*bgp.BGPNotification).ErrorCode == bgp.BGP_ERROR_HOLD_TIMER_EXPIRED) ||
					err.Type == fsmReadFailed ||
					err.Type == fsmWriteFailed {
					err = *newfsmStateReason(fsmGracefulRestart, nil, nil)
					log.WithFields(log.Fields{
						"Topic": "Peer",
						"Key":   fsm.pConf.State.NeighborAddress,
						"State": fsm.state.String(),
					}).Info("peer graceful restart")
					fsm.gracefulRestartTimer.Reset(time.Duration(fsm.pConf.GracefulRestart.State.PeerRestartTime) * time.Second)
				}
			}
			fsm.lock.RUnlock()
			return bgp.BGP_FSM_IDLE, &err
		case <-holdTimer.C:
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
			}).Warn("hold timer expired")
			fsm.lock.RUnlock()
			m := bgp.NewBGPNotificationMessage(bgp.BGP_ERROR_HOLD_TIMER_EXPIRED, 0, nil)
			h.outgoing.In() <- &fsmOutgoingMsg{Notification: m}
			fsm.lock.RLock()
			s := fsm.pConf.GracefulRestart.State
			fsm.lock.RUnlock()
			// Do not return hold timer expired to server if graceful restart is enabled
			// Let it fallback to read/write error or fsmNotificationSent handled above
			// Reference: https://github.com/osrg/gobgp/issues/2174
			if !s.Enabled {
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmHoldTimerExpired, m, nil)
			}
		case <-h.holdTimerResetCh:
			fsm.lock.RLock()
			if fsm.pConf.Timers.State.NegotiatedHoldTime != 0 {
				holdTimer.Reset(time.Second * time.Duration(fsm.pConf.Timers.State.NegotiatedHoldTime))
			}
			fsm.lock.RUnlock()
			/*case stateOp := <-fsm.adminStateCh:
			err := h.changeadminState(stateOp.State)
			if err == nil {
				switch stateOp.State {
				case adminStateDown:
					m := bgp.NewBGPNotificationMessage(bgp.BGP_ERROR_CEASE, bgp.BGP_ERROR_SUB_ADMINISTRATIVE_SHUTDOWN, stateOp.Communication)
					h.outgoing.In() <- &fsmOutgoingMsg{Notification: m}
				}
			}*/
		}
	}
}

type fsmOutgoingMsg struct {
	Paths        []*table.Path
	Notification *bgp.BGPMessage
	StayIdle     bool
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
	send := func(m *bgp.BGPMessage) error {
		fsm.lock.RLock()
		if fsm.twoByteAsTrans && m.Header.Type == bgp.BGP_MSG_UPDATE {
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
				"Data":  m,
			}).Debug("update for 2byte AS peer")
			table.UpdatePathAttrs2ByteAs(m.Body.(*bgp.BGPUpdate))
			table.UpdatePathAggregator2ByteAs(m.Body.(*bgp.BGPUpdate))
		}
		b, err := m.Serialize(h.fsm.marshallingOptions)
		fsm.lock.RUnlock()
		if err != nil {
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
				"Data":  err,
			}).Warn("failed to serialize")
			fsm.lock.RUnlock()
			fsm.bgpMessageStateUpdate(0, false)
			return nil
		}
		fsm.lock.RLock()
		err = conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(fsm.pConf.Timers.State.NegotiatedHoldTime)))
		fsm.lock.RUnlock()
		if err != nil {
			sendToStateReasonCh(fsmWriteFailed, nil)
			conn.Close()
			return fmt.Errorf("failed to set write deadline")
		}
		_, err = conn.Write(b)
		if err != nil {
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
				"Data":  err,
			}).Warn("failed to send")
			fsm.lock.RUnlock()
			sendToStateReasonCh(fsmWriteFailed, nil)
			conn.Close()
			return fmt.Errorf("closed")
		}
		fsm.bgpMessageStateUpdate(m.Header.Type, false)

		switch m.Header.Type {
		case bgp.BGP_MSG_NOTIFICATION:
			body := m.Body.(*bgp.BGPNotification)
			if body.ErrorCode == bgp.BGP_ERROR_CEASE && (body.ErrorSubcode == bgp.BGP_ERROR_SUB_ADMINISTRATIVE_SHUTDOWN || body.ErrorSubcode == bgp.BGP_ERROR_SUB_ADMINISTRATIVE_RESET) {
				communication, rest := decodeAdministrativeCommunication(body.Data)
				fsm.lock.RLock()
				log.WithFields(log.Fields{
					"Topic":               "Peer",
					"Key":                 fsm.pConf.State.NeighborAddress,
					"State":               fsm.state.String(),
					"Code":                body.ErrorCode,
					"Subcode":             body.ErrorSubcode,
					"Communicated-Reason": communication,
					"Data":                rest,
				}).Warn("sent notification")
				fsm.lock.RUnlock()
			} else {
				fsm.lock.RLock()
				log.WithFields(log.Fields{
					"Topic":   "Peer",
					"Key":     fsm.pConf.State.NeighborAddress,
					"State":   fsm.state.String(),
					"Code":    body.ErrorCode,
					"Subcode": body.ErrorSubcode,
					"Data":    body.Data,
				}).Warn("sent notification")
				fsm.lock.RUnlock()
			}
			sendToStateReasonCh(fsmNotificationSent, m)
			conn.Close()
			return fmt.Errorf("closed")
		case bgp.BGP_MSG_UPDATE:
			update := m.Body.(*bgp.BGPUpdate)
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic":       "Peer",
				"Key":         fsm.pConf.State.NeighborAddress,
				"State":       fsm.state.String(),
				"nlri":        update.NLRI,
				"withdrawals": update.WithdrawnRoutes,
				"attributes":  update.PathAttributes,
			}).Debug("sent update")
			fsm.lock.RUnlock()
		default:
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
				"data":  m,
			}).Debug("sent")
			fsm.lock.RUnlock()
		}
		return nil
	}

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
