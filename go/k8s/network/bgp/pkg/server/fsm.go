package server

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"strconv"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/packet/bgp"
	"k8s-lx1036/k8s/network/bgp/pkg/table"

	log "github.com/sirupsen/logrus"
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

type fsmOutgoingMsg struct {
	Paths        []*table.Path
	Notification *bgp.BGPMessage
	StayIdle     bool
}

const (
	holdtimeOpensent = 240
	holdtimeIdle     = 5
)

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

// INFO: fsm state
//  idle -> active -> opensent -> openconfirm -> established
//  idle:
//  active: 与 route server 建立 tcp connection
//  opensent: 按照 BGP 标准，发送 opensent message 给 router server，然后接收消息，最后再发送 keepalive 心跳信息ack下
//  openconfirm:

type fsm struct {
	wg *sync.WaitGroup

	gConf *config.Global
	pConf *config.Neighbor
	lock  sync.RWMutex
	state bgp.FSMState

	outgoingCh chan *fsmOutgoingMsg
	incomingCh chan *fsmMsg

	reason           *fsmStateReason
	conn             net.Conn
	connCh           chan net.Conn
	idleHoldTime     float64
	opensentHoldTime float64
	adminState       adminState
	adminStateCh     chan adminStateOperation
	//h                    *fsmHandler
	rfMap                map[bgp.RouteFamily]bgp.BGPAddPathMode
	capMap               map[bgp.BGPCapabilityCode][]bgp.ParameterCapabilityInterface
	recvOpen             *bgp.BGPMessage
	peerInfo             *table.PeerInfo
	gracefulRestartTimer *time.Timer
	twoByteAsTrans       bool
	marshallingOptions   *bgp.MarshallingOption
	notification         chan *bgp.BGPMessage

	// fsm handler
	sentNotification *bgp.BGPMessage
	stateReasonCh    chan fsmStateReason
	holdTimerResetCh chan bool
	msgCh            chan *fsmMsg
}

func newFSM(gConf *config.Global, pConf *config.Neighbor) *fsm {
	adminState := adminStateUp
	if pConf.Config.AdminDown {
		adminState = adminStateDown
	}
	pConf.State.SessionState = config.IntToSessionStateMap[int(bgp.BGP_FSM_IDLE)]
	pConf.Timers.State.Downtime = time.Now().Unix()
	fsm := &fsm{
		wg: &sync.WaitGroup{},

		outgoingCh: make(chan *fsmOutgoingMsg, 1024),
		msgCh:      make(chan *fsmMsg, 1024),
		//incomingCh: make(chan *fsmMsg, 1024), // 不要这里实例化，在 server 上层实例化

		gConf:                gConf,
		pConf:                pConf,
		state:                bgp.BGP_FSM_IDLE,
		connCh:               make(chan net.Conn, 1),
		opensentHoldTime:     float64(holdtimeOpensent),
		adminState:           adminState,
		adminStateCh:         make(chan adminStateOperation, 1),
		rfMap:                make(map[bgp.RouteFamily]bgp.BGPAddPathMode),
		capMap:               make(map[bgp.BGPCapabilityCode][]bgp.ParameterCapabilityInterface),
		peerInfo:             table.NewPeerInfo(gConf, pConf),
		gracefulRestartTimer: time.NewTimer(time.Hour),

		// fsm handler
		notification:     make(chan *bgp.BGPMessage, 1),
		stateReasonCh:    make(chan fsmStateReason, 2),
		holdTimerResetCh: make(chan bool, 2),
	}
	fsm.gracefulRestartTimer.Stop()
	return fsm
}

func (fsm *fsm) start(ctx context.Context) error {
	oldState := fsm.state
	nextState := bgp.FSMState(-1)
	fsmState := fsm.state
	var reason *fsmStateReason

	switch fsmState {
	case bgp.BGP_FSM_IDLE:
		nextState, reason = fsm.idle(ctx) // idle -> active
	case bgp.BGP_FSM_ACTIVE:
		nextState, reason = fsm.active(ctx) // 在 active state 和交换机建立 tcp connection
	case bgp.BGP_FSM_OPENSENT:
		nextState, reason = fsm.opensent(ctx)
	case bgp.BGP_FSM_OPENCONFIRM:
		nextState, reason = fsm.openconfirm(ctx)
	case bgp.BGP_FSM_ESTABLISHED:
		nextState, reason = fsm.established(ctx)
	}

	fsm.reason = reason
	if nextState == bgp.BGP_FSM_ESTABLISHED && oldState == bgp.BGP_FSM_OPENCONFIRM {
		klog.Infof(fmt.Sprintf("[fsm loop]peer %s state is %s", fsm.pConf.State.NeighborAddress, fsm.state.String()))
	}
	if oldState == bgp.BGP_FSM_ESTABLISHED {
		if fsm.sentNotification != nil {
			fsm.reason.Type = fsmNotificationSent
			fsm.reason.BGPNotification = fsm.sentNotification
		}
		klog.Infof(fmt.Sprintf("[fsm loop]peer %s is down from established to %s: %s",
			fsm.pConf.State.NeighborAddress, fsm.state.String(), fsm.reason.String()))
	}

	fsm.lock.RLock()
	fsm.incomingCh <- &fsmMsg{ // idle -> active
		fsm:         fsm,
		MsgType:     fsmMsgStateChange,
		MsgSrc:      fsm.pConf.State.NeighborAddress,
		MsgData:     nextState,
		StateReason: reason,
	}
	fsm.lock.RUnlock()
	return nil
}

func (fsm *fsm) idle(ctx context.Context) (bgp.FSMState, *fsmStateReason) {
	idleHoldTimer := time.NewTimer(time.Second * time.Duration(fsm.idleHoldTime)) // INFO: 起始为0，<-idleHoldTimer.C 会先走
	for {
		select {
		case <-ctx.Done():
			return -1, newfsmStateReason(fsmDying, nil, nil)
		case <-fsm.gracefulRestartTimer.C:
			if fsm.pConf.GracefulRestart.State.PeerRestarting {
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmRestartTimerExpired, nil, nil)
			}

		case conn, ok := <-fsm.connCh:
			if !ok {
				break
			}
			conn.Close()
			klog.Infof("Closed an accepted connection")

		case <-idleHoldTimer.C:
			if fsm.adminState == adminStateUp {
				fsm.idleHoldTime = holdtimeIdle // 5s 周期
				klog.Infof("IdleHoldTimer expired")
				return bgp.BGP_FSM_ACTIVE, newfsmStateReason(fsmIdleTimerExpired, nil, nil)
			} else {
				klog.Infof("IdleHoldTimer expired, but stay at idle because the admin state is DOWN")
			}

		case stateOp := <-fsm.adminStateCh:
			err := fsm.changeadminState(stateOp.State)
			if err == nil {
				switch stateOp.State {
				case adminStateDown:
					// stop idle hold timer
					idleHoldTimer.Stop()

				case adminStateUp:
					// restart idle hold timer
					fsm.lock.RLock()
					idleHoldTimer.Reset(time.Second * time.Duration(fsm.idleHoldTime))
					fsm.lock.RUnlock()
				}
			}
		}
	}
}

func (fsm *fsm) active(ctx context.Context) (bgp.FSMState, *fsmStateReason) {
	// try to connect router server
	if !fsm.pConf.Transport.Config.PassiveMode {
		go func() {
			laddr, _ := net.ResolveTCPAddr("tcp", net.JoinHostPort(fsm.pConf.Transport.Config.LocalAddress, "0"))
			addr := fsm.pConf.State.NeighborAddress
			port := int(bgp.BGP_PORT)
			if fsm.pConf.Transport.Config.RemotePort != 0 {
				port = int(fsm.pConf.Transport.Config.RemotePort)
			}
			d := net.Dialer{
				LocalAddr: laddr,
				Timeout:   5 * time.Second,
			}
			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
			if err != nil {
				klog.Fatalf(fmt.Sprintf("[fsm active]conn err:%v", err))
			}
			fsm.connCh <- conn
		}()
	}

	for {
		select {
		case <-ctx.Done():
			return -1, newfsmStateReason(fsmDying, nil, nil)
		case conn, ok := <-fsm.connCh:
			if !ok {
				break
			}
			fsm.lock.Lock()
			fsm.conn = conn
			fsm.lock.Unlock()
			// we don't implement delayed open timer so move to opensent right away.
			return bgp.BGP_FSM_OPENSENT, newfsmStateReason(fsmNewConnection, nil, nil)
		case <-fsm.gracefulRestartTimer.C:
			fsm.lock.RLock()
			restarting := fsm.pConf.GracefulRestart.State.PeerRestarting
			fsm.lock.RUnlock()
			if restarting {
				fsm.lock.RLock()
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Key":   fsm.pConf.State.NeighborAddress,
					"State": fsm.state.String(),
				}).Warn("graceful restart timer expired")
				fsm.lock.RUnlock()
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmRestartTimerExpired, nil, nil)
			}
		case err := <-fsm.stateReasonCh:
			return bgp.BGP_FSM_IDLE, &err
		case stateOp := <-fsm.adminStateCh:
			err := fsm.changeadminState(stateOp.State)
			if err == nil {
				switch stateOp.State {
				case adminStateDown:
					return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmAdminDown, nil, nil)
				case adminStateUp:
					log.WithFields(log.Fields{
						"Topic":      "Peer",
						"Key":        fsm.pConf.State.NeighborAddress,
						"State":      fsm.state.String(),
						"adminState": stateOp.State.String(),
					}).Panic("code logic bug")
				}
			}
		}
	}
}

// INFO: 按照 BGP 标准，发送 opensent message 给 router server，然后接收消息，最后再发送 keepalive 心跳信息ack下
func (fsm *fsm) opensent(ctx context.Context) (bgp.FSMState, *fsmStateReason) {
	m := buildopen(fsm.gConf, fsm.pConf)
	b, _ := m.Serialize()
	fsm.conn.Write(b)
	fsm.bgpMessageStateUpdate(m.Header.Type, false)

	go fsm.recvMessage()

	holdTimer := time.NewTimer(time.Second * time.Duration(fsm.opensentHoldTime)) // 4 min
	for {
		select {
		case <-ctx.Done():
			fsm.conn.Close()
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
		case <-fsm.gracefulRestartTimer.C:
			fsm.lock.RLock()
			restarting := fsm.pConf.GracefulRestart.State.PeerRestarting
			fsm.lock.RUnlock()
			if restarting {
				fsm.lock.RLock()
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Key":   fsm.pConf.State.NeighborAddress,
					"State": fsm.state.String(),
				}).Warn("graceful restart timer expired")
				fsm.lock.RUnlock()
				fsm.conn.Close()
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmRestartTimerExpired, nil, nil)
			}
		case msg := <-fsm.msgCh:
			switch m := msg.MsgData.(type) {
			case *bgp.BGPMessage:
				if m.Header.Type == bgp.BGP_MSG_OPEN {
					fsm.recvOpen = m
					body := m.Body.(*bgp.BGPOpen)
					fsmPeerAS := fsm.pConf.Config.PeerAs
					peerAs, err := bgp.ValidateOpenMsg(body, fsmPeerAS, fsm.peerInfo.LocalAS, net.ParseIP(fsm.gConf.Config.RouterId))
					if err != nil {
						m, _ := fsm.sendNotificationFromErrorMsg(err.(*bgp.MessageError))
						return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmInvalidMsg, m, nil)
					}

					fsm.pConf.State.PeerType = fsm.pConf.Config.PeerType

					fsm.lock.Lock()
					fsm.pConf.State.PeerAs = peerAs
					fsm.peerInfo.AS = peerAs
					fsm.peerInfo.ID = body.ID
					fsm.capMap, fsm.rfMap = open2Cap(body, fsm.pConf)
					if _, y := fsm.capMap[bgp.BGP_CAP_ADD_PATH]; y {
						fsm.marshallingOptions = &bgp.MarshallingOption{
							AddPath: fsm.rfMap,
						}
					} else {
						fsm.marshallingOptions = nil
					}

					// calculate HoldTime
					// RFC 4271 P.13
					// a BGP speaker MUST calculate the value of the Hold Timer
					// by using the smaller of its configured Hold Time and the Hold Time
					// received in the OPEN message.
					holdTime := float64(body.HoldTime)
					myHoldTime := fsm.pConf.Timers.Config.HoldTime
					if holdTime > myHoldTime {
						fsm.pConf.Timers.State.NegotiatedHoldTime = myHoldTime
					} else {
						fsm.pConf.Timers.State.NegotiatedHoldTime = holdTime
					}

					keepalive := fsm.pConf.Timers.Config.KeepaliveInterval
					if n := fsm.pConf.Timers.State.NegotiatedHoldTime; n < myHoldTime {
						keepalive = n / 3
					}
					fsm.pConf.Timers.State.KeepaliveInterval = keepalive

					gr, ok := fsm.capMap[bgp.BGP_CAP_GRACEFUL_RESTART]
					if fsm.pConf.GracefulRestart.Config.Enabled && ok {
						state := &fsm.pConf.GracefulRestart.State
						state.Enabled = true
						cap := gr[len(gr)-1].(*bgp.CapGracefulRestart)
						state.PeerRestartTime = uint16(cap.Time)

						for _, t := range cap.Tuples {
							n := bgp.AddressFamilyNameMap[bgp.AfiSafiToRouteFamily(t.AFI, t.SAFI)]
							for i, a := range fsm.pConf.AfiSafis {
								if string(a.Config.AfiSafiName) == n {
									fsm.pConf.AfiSafis[i].MpGracefulRestart.State.Enabled = true
									fsm.pConf.AfiSafis[i].MpGracefulRestart.State.Received = true
									break
								}
							}
						}

						// RFC 4724 4.1
						// To re-establish the session with its peer, the Restarting Speaker
						// MUST set the "Restart State" bit in the Graceful Restart Capability
						// of the OPEN message.
						if fsm.pConf.GracefulRestart.State.PeerRestarting && cap.Flags&0x08 == 0 {
							log.WithFields(log.Fields{
								"Topic": "Peer",
								"Key":   fsm.pConf.State.NeighborAddress,
								"State": fsm.state.String(),
							}).Warn("restart flag is not set")
							// just ignore
						}

						// RFC 4724 3
						// The most significant bit is defined as the Restart State (R)
						// bit, ...(snip)... When set (value 1), this bit
						// indicates that the BGP speaker has restarted, and its peer MUST
						// NOT wait for the End-of-RIB marker from the speaker before
						// advertising routing information to the speaker.
						if fsm.pConf.GracefulRestart.State.LocalRestarting && cap.Flags&0x08 != 0 {
							log.WithFields(log.Fields{
								"Topic": "Peer",
								"Key":   fsm.pConf.State.NeighborAddress,
								"State": fsm.state.String(),
							}).Debug("peer has restarted, skipping wait for EOR")
							for i := range fsm.pConf.AfiSafis {
								fsm.pConf.AfiSafis[i].MpGracefulRestart.State.EndOfRibReceived = true
							}
						}
						if fsm.pConf.GracefulRestart.Config.NotificationEnabled && cap.Flags&0x04 > 0 {
							fsm.pConf.GracefulRestart.State.NotificationEnabled = true
						}
					}
					llgr, ok2 := fsm.capMap[bgp.BGP_CAP_LONG_LIVED_GRACEFUL_RESTART]
					if fsm.pConf.GracefulRestart.Config.LongLivedEnabled && ok && ok2 {
						fsm.pConf.GracefulRestart.State.LongLivedEnabled = true
						cap := llgr[len(llgr)-1].(*bgp.CapLongLivedGracefulRestart)
						for _, t := range cap.Tuples {
							n := bgp.AddressFamilyNameMap[bgp.AfiSafiToRouteFamily(t.AFI, t.SAFI)]
							for i, a := range fsm.pConf.AfiSafis {
								if string(a.Config.AfiSafiName) == n {
									fsm.pConf.AfiSafis[i].LongLivedGracefulRestart.State.Enabled = true
									fsm.pConf.AfiSafis[i].LongLivedGracefulRestart.State.Received = true
									fsm.pConf.AfiSafis[i].LongLivedGracefulRestart.State.PeerRestartTime = t.RestartTime
									break
								}
							}
						}
					}

					fsm.lock.Unlock()

					// INFO: 给 router server 发送心跳，确保 ack
					msg := bgp.NewBGPKeepAliveMessage()
					b, _ := msg.Serialize()
					fsm.conn.Write(b)
					fsm.bgpMessageStateUpdate(msg.Header.Type, false)
					return bgp.BGP_FSM_OPENCONFIRM, newfsmStateReason(fsmOpenMsgReceived, nil, nil)
				} else {
					// send notification?
					fsm.conn.Close()
					return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmInvalidMsg, nil, nil)
				}
			case *bgp.MessageError:
				msg, _ := fsm.sendNotificationFromErrorMsg(m)
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmInvalidMsg, msg, nil)
			default:
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Key":   fsm.pConf.State.NeighborAddress,
					"State": fsm.state.String(),
					"Data":  msg.MsgData,
				}).Panic("unknown msg type")
			}
		case err := <-fsm.stateReasonCh:
			fsm.conn.Close()
			return bgp.BGP_FSM_IDLE, &err
		case <-holdTimer.C:
			m, _ := fsm.sendNotification(bgp.BGP_ERROR_HOLD_TIMER_EXPIRED, 0, nil, "hold timer expired")
			return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmHoldTimerExpired, m, nil)
		case stateOp := <-fsm.adminStateCh:
			err := fsm.changeadminState(stateOp.State)
			if err == nil {
				switch stateOp.State {
				case adminStateDown:
					fsm.conn.Close()
					return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmAdminDown, m, nil)
				case adminStateUp:
					log.WithFields(log.Fields{
						"Topic":      "Peer",
						"Key":        fsm.pConf.State.NeighborAddress,
						"State":      fsm.state.String(),
						"adminState": stateOp.State.String(),
					}).Panic("code logic bug")
				}
			}
		}
	}
}

func (fsm *fsm) openconfirm(ctx context.Context) (bgp.FSMState, *fsmStateReason) {
	ticker := keepaliveTicker(fsm)

	// INFO: 从 router server 收到心跳 ack 信息，下一步则是 established state
	go fsm.recvMessage()

	for {
		select {
		case <-ctx.Done():
			fsm.conn.Close()
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
		case <-fsm.gracefulRestartTimer.C:
			fsm.lock.RLock()
			restarting := fsm.pConf.GracefulRestart.State.PeerRestarting
			fsm.lock.RUnlock()
			if restarting {
				fsm.lock.RLock()
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Key":   fsm.pConf.State.NeighborAddress,
					"State": fsm.state.String(),
				}).Warn("graceful restart timer expired")
				fsm.lock.RUnlock()
				fsm.conn.Close()
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmRestartTimerExpired, nil, nil)
			}
		case <-ticker.C: // INFO: 同时每秒发送心跳keepalive信息给 router server
			m := bgp.NewBGPKeepAliveMessage()
			b, _ := m.Serialize()
			fsm.conn.Write(b)
			fsm.bgpMessageStateUpdate(m.Header.Type, false)
		case msg := <-fsm.msgCh:
			switch m := msg.MsgData.(type) {
			case *bgp.BGPMessage:
				if m.Header.Type == bgp.BGP_MSG_KEEPALIVE {
					return bgp.BGP_FSM_ESTABLISHED, newfsmStateReason(fsmOpenMsgNegotiated, nil, nil)
				}
				// send notification ?
				fsm.conn.Close()
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmInvalidMsg, nil, nil)
			case *bgp.MessageError:
				msg, _ := fsm.sendNotificationFromErrorMsg(m)
				return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmInvalidMsg, msg, nil)
			default:
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Key":   fsm.pConf.State.NeighborAddress,
					"State": fsm.state.String(),
					"Data":  msg.MsgData,
				}).Panic("unknown msg type")
			}
		case err := <-fsm.stateReasonCh:
			fsm.conn.Close()
			return bgp.BGP_FSM_IDLE, &err
		case stateOp := <-fsm.adminStateCh:
			err := fsm.changeadminState(stateOp.State)
			if err == nil {
				switch stateOp.State {
				case adminStateDown:
					fsm.conn.Close()
					return bgp.BGP_FSM_IDLE, newfsmStateReason(fsmAdminDown, nil, nil)
				case adminStateUp:
					log.WithFields(log.Fields{
						"Topic":      "Peer",
						"Key":        fsm.pConf.State.NeighborAddress,
						"State":      fsm.state.String(),
						"adminState": stateOp.State.String(),
					}).Panic("code logic bug")
				}
			}
		}
	}
}

// INFO: 建立 establish 之后，则周期发送心跳信息
func (fsm *fsm) established(ctx context.Context) (bgp.FSMState, *fsmStateReason) {
	var wg sync.WaitGroup
	fsm.lock.Lock()
	fsm.lock.Unlock()

	defer wg.Wait()
	wg.Add(2)

	go fsm.sendMessageloop(ctx)
	go func() {
		for {
			fmsg, err := fsm.recvMessageWithError()
			if fmsg != nil {
				fsm.incomingCh <- fmsg
			}
			if err != nil {
				klog.Errorf(fmt.Sprintf("[fsm establish]receive msg err:%v", err))
				return
			}
		}
	}()

	fsm.gracefulRestartTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			select {
			case m := <-fsm.notification:
				b, _ := m.Serialize(fsm.marshallingOptions)
				fsm.conn.Write(b)
			default:
				// nothing to do
			}
			fsm.conn.Close()
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
		case err := <-fsm.stateReasonCh:
			fsm.conn.Close()
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
		case stateOp := <-fsm.adminStateCh:
			err := fsm.changeadminState(stateOp.State)
			if err == nil {
				switch stateOp.State {
				case adminStateDown:
					m := bgp.NewBGPNotificationMessage(bgp.BGP_ERROR_CEASE, bgp.BGP_ERROR_SUB_ADMINISTRATIVE_SHUTDOWN, stateOp.Communication)
					fsm.outgoingCh <- &fsmOutgoingMsg{Notification: m}
				}
			}
		}
	}
}

func (fsm *fsm) sendMessageloop(ctx context.Context) error {
	sendToStateReasonCh := func(typ fsmStateReasonType, notif *bgp.BGPMessage) {
		// probably doesn't happen but be cautious
		select {
		case fsm.stateReasonCh <- *newfsmStateReason(typ, notif, nil):
		default:
		}
	}

	conn := fsm.conn
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
		b, err := m.Serialize(fsm.marshallingOptions)
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
		klog.Infof(fmt.Sprintf("[sendMessageloop]send msg %s to router", b))

		err = conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(fsm.pConf.Timers.State.NegotiatedHoldTime)))
		if err != nil {
			sendToStateReasonCh(fsmWriteFailed, nil)
			conn.Close()
			return fmt.Errorf("failed to set write deadline")
		}
		_, err = conn.Write(b)
		if err != nil {
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
		case m := <-fsm.outgoingCh:
			fsm.lock.RLock()
			options := fsm.marshallingOptions
			fsm.lock.RUnlock()
			for _, msg := range table.CreateUpdateMsgFromPaths(m.Paths, options) {
				if err := send(msg); err != nil {
					return nil
				}
			}
			if m.Notification != nil {
				if m.StayIdle {
					// current user is only prefix-limit
					// fix me if this is not the case
					fsm.changeadminState(adminStatePfxCt)
				}
				if err := send(m.Notification); err != nil {
					return nil
				}
			}
		case <-ticker.C:
			if err := send(bgp.NewBGPKeepAliveMessage()); err != nil {
				return nil
			}
		}
	}
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

func (fsm *fsm) StateChange(nextState bgp.FSMState) {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()

	klog.Infof(fmt.Sprintf("[StateChange]state changed, key:%s, old:%s, new:%s, reason:%s",
		fsm.pConf.State.NeighborAddress, fsm.state.String(), nextState.String(), fsm.reason))
	fsm.state = nextState
	switch nextState {
	case bgp.BGP_FSM_ESTABLISHED:
		fsm.pConf.Timers.State.Uptime = time.Now().Unix()
		fsm.pConf.State.EstablishedCount++
		// reset the state set by the previous session
		fsm.twoByteAsTrans = false
		if _, y := fsm.capMap[bgp.BGP_CAP_FOUR_OCTET_AS_NUMBER]; !y {
			fsm.twoByteAsTrans = true
			break
		}
		y := func() bool {
			for _, c := range capabilitiesFromConfig(fsm.pConf) {
				switch c.(type) {
				case *bgp.CapFourOctetASNumber:
					return true
				}
			}
			return false
		}()
		if !y {
			fsm.twoByteAsTrans = true
		}
	default:
		fsm.pConf.Timers.State.Downtime = time.Now().Unix()
	}
}

func (fsm *fsm) RemoteHostPort() (string, uint16) {
	return hostport(fsm.conn.RemoteAddr())

}

func (fsm *fsm) LocalHostPort() (string, uint16) {
	return hostport(fsm.conn.LocalAddr())
}

func (fsm *fsm) sendNotificationFromErrorMsg(e *bgp.MessageError) (*bgp.BGPMessage, error) {
	if fsm.conn != nil {
		m := bgp.NewBGPNotificationMessage(e.TypeCode, e.SubTypeCode, e.Data)
		b, _ := m.Serialize()
		_, err := fsm.conn.Write(b)
		if err == nil {
			fsm.bgpMessageStateUpdate(m.Header.Type, false)
			fsm.sentNotification = m
		}
		fsm.conn.Close()
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   fsm.pConf.State.NeighborAddress,
			"Data":  e,
		}).Warn("sent notification")
		return m, nil
	}
	return nil, fmt.Errorf("can't send notification to %s since TCP connection is not established", fsm.pConf.State.NeighborAddress)
}

func (fsm *fsm) sendNotification(code, subType uint8, data []byte, msg string) (*bgp.BGPMessage, error) {
	e := bgp.NewMessageError(code, subType, data, msg)
	return fsm.sendNotificationFromErrorMsg(e.(*bgp.MessageError))
}

func (fsm *fsm) afiSafiDisable(rf bgp.RouteFamily) string {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()

	n := bgp.AddressFamilyNameMap[rf]

	for i, a := range fsm.pConf.AfiSafis {
		if string(a.Config.AfiSafiName) == n {
			fsm.pConf.AfiSafis[i].State.Enabled = false
			break
		}
	}
	newList := make([]bgp.ParameterCapabilityInterface, 0)
	for _, c := range fsm.capMap[bgp.BGP_CAP_MULTIPROTOCOL] {
		if c.(*bgp.CapMultiProtocol).CapValue == rf {
			continue
		}
		newList = append(newList, c)
	}
	fsm.capMap[bgp.BGP_CAP_MULTIPROTOCOL] = newList
	return n
}

func (fsm *fsm) handlingError(m *bgp.BGPMessage, e error, useRevisedError bool) bgp.ErrorHandling {
	// ineffectual assignment to handling (ineffassign)
	var handling bgp.ErrorHandling
	if m.Header.Type == bgp.BGP_MSG_UPDATE && useRevisedError {
		factor := e.(*bgp.MessageError)
		handling = factor.ErrorHandling
		switch handling {
		case bgp.ERROR_HANDLING_ATTRIBUTE_DISCARD:
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
				"error": e,
			}).Warn("Some attributes were discarded")
			fsm.lock.RUnlock()
		case bgp.ERROR_HANDLING_TREAT_AS_WITHDRAW:
			m.Body = bgp.TreatAsWithdraw(m.Body.(*bgp.BGPUpdate))
			fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
				"error": e,
			}).Warn("the received Update message was treated as withdraw")
			fsm.lock.RUnlock()
		case bgp.ERROR_HANDLING_AFISAFI_DISABLE:
			rf := extractRouteFamily(factor.ErrorAttribute)
			if rf == nil {
				fsm.lock.RLock()
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Key":   fsm.pConf.State.NeighborAddress,
					"State": fsm.state.String(),
				}).Warn("Error occurred during AFI/SAFI disabling")
				fsm.lock.RUnlock()
			} else {
				n := fsm.afiSafiDisable(*rf)
				fsm.lock.RLock()
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Key":   fsm.pConf.State.NeighborAddress,
					"State": fsm.state.String(),
					"error": e,
				}).Warnf("Capability %s was disabled", n)
				fsm.lock.RUnlock()
			}
		}
	} else {
		handling = bgp.ERROR_HANDLING_SESSION_RESET
	}
	return handling
}

func (fsm *fsm) recvMessageWithError() (*fsmMsg, error) {
	sendToStateReasonCh := func(typ fsmStateReasonType, notif *bgp.BGPMessage) {
		// probably doesn't happen but be cautious
		select {
		case fsm.stateReasonCh <- *newfsmStateReason(typ, notif, nil):
		default:
		}
	}

	headerBuf, err := readAll(fsm.conn, bgp.BGP_HEADER_LENGTH)
	if err != nil {
		sendToStateReasonCh(fsmReadFailed, nil)
		return nil, err
	}

	hd := &bgp.BGPHeader{}
	err = hd.DecodeFromBytes(headerBuf)
	if err != nil {
		fsm.bgpMessageStateUpdate(0, true)
		fsm.lock.RLock()
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   fsm.pConf.State.NeighborAddress,
			"State": fsm.state.String(),
			"error": err,
		}).Warn("Session will be reset due to malformed BGP Header")
		fmsg := &fsmMsg{
			fsm:     fsm,
			MsgType: fsmMsgBGPMessage,
			MsgSrc:  fsm.pConf.State.NeighborAddress,
			MsgData: err,
		}
		fsm.lock.RUnlock()
		return fmsg, err
	}

	bodyBuf, err := readAll(fsm.conn, int(hd.Len)-bgp.BGP_HEADER_LENGTH)
	if err != nil {
		sendToStateReasonCh(fsmReadFailed, nil)
		return nil, err
	}

	now := time.Now()
	handling := bgp.ERROR_HANDLING_NONE

	fsm.lock.RLock()
	useRevisedError := fsm.pConf.ErrorHandling.Config.TreatAsWithdraw
	options := fsm.marshallingOptions
	fsm.lock.RUnlock()

	m, err := bgp.ParseBGPBody(hd, bodyBuf, options)
	if err != nil {
		handling = fsm.handlingError(m, err, useRevisedError)
		fsm.bgpMessageStateUpdate(0, true)
	} else {
		fsm.bgpMessageStateUpdate(m.Header.Type, true)
		err = bgp.ValidateBGPMessage(m)
	}
	fsm.lock.RLock()
	fmsg := &fsmMsg{
		fsm:       fsm,
		MsgType:   fsmMsgBGPMessage,
		MsgSrc:    fsm.pConf.State.NeighborAddress,
		timestamp: now,
	}
	fsm.lock.RUnlock()

	switch handling {
	case bgp.ERROR_HANDLING_AFISAFI_DISABLE:
		fmsg.MsgData = m
		return fmsg, nil
	case bgp.ERROR_HANDLING_SESSION_RESET:
		fsm.lock.RLock()
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   fsm.pConf.State.NeighborAddress,
			"State": fsm.state.String(),
			"error": err,
		}).Warn("Session will be reset due to malformed BGP message")
		fsm.lock.RUnlock()
		fmsg.MsgData = err
		return fmsg, err
	default:
		fmsg.MsgData = m

		fsm.lock.RLock()
		establishedState := fsm.state == bgp.BGP_FSM_ESTABLISHED
		fsm.lock.RUnlock()

		if establishedState {
			switch m.Header.Type {
			case bgp.BGP_MSG_ROUTE_REFRESH:
				fmsg.MsgType = fsmMsgRouteRefresh
			case bgp.BGP_MSG_UPDATE:
				// if the length of fsm.holdTimerResetCh
				// isn't zero, the timer will be reset
				// soon anyway.
				select {
				case fsm.holdTimerResetCh <- true:
				default:
				}
				body := m.Body.(*bgp.BGPUpdate)
				isEBGP := fsm.pConf.IsEBGPPeer(fsm.gConf)
				isConfed := fsm.pConf.IsConfederationMember(fsm.gConf)

				fmsg.payload = make([]byte, len(headerBuf)+len(bodyBuf))
				copy(fmsg.payload, headerBuf)
				copy(fmsg.payload[len(headerBuf):], bodyBuf)

				fsm.lock.RLock()
				rfMap := fsm.rfMap
				fsm.lock.RUnlock()
				ok, err := bgp.ValidateUpdateMsg(body, rfMap, isEBGP, isConfed)
				if !ok {
					handling = fsm.handlingError(m, err, useRevisedError)
				}
				if handling == bgp.ERROR_HANDLING_SESSION_RESET {
					fsm.lock.RLock()
					log.WithFields(log.Fields{
						"Topic": "Peer",
						"Key":   fsm.pConf.State.NeighborAddress,
						"State": fsm.state.String(),
						"error": err,
					}).Warn("Session will be reset due to malformed BGP update message")
					fsm.lock.RUnlock()
					fmsg.MsgData = err
					return fmsg, err
				}

				table.UpdatePathAttrs4ByteAs(body)
				if err = table.UpdatePathAggregator4ByteAs(body); err != nil {
					fmsg.MsgData = err
					return fmsg, err
				}

				fsm.lock.RLock()
				peerInfo := fsm.peerInfo
				fsm.lock.RUnlock()
				fmsg.PathList = table.ProcessMessage(m, peerInfo, fmsg.timestamp)
				fallthrough
			case bgp.BGP_MSG_KEEPALIVE:
				// if the length of fsm.holdTimerResetCh
				// isn't zero, the timer will be reset
				// soon anyway.
				select {
				case fsm.holdTimerResetCh <- true:
				default:
				}
				if m.Header.Type == bgp.BGP_MSG_KEEPALIVE {
					return nil, nil
				}
			case bgp.BGP_MSG_NOTIFICATION:
				body := m.Body.(*bgp.BGPNotification)
				if body.ErrorCode == bgp.BGP_ERROR_CEASE && (body.ErrorSubcode == bgp.BGP_ERROR_SUB_ADMINISTRATIVE_SHUTDOWN || body.ErrorSubcode == bgp.BGP_ERROR_SUB_ADMINISTRATIVE_RESET) {
					communication, rest := decodeAdministrativeCommunication(body.Data)
					fsm.lock.RLock()
					log.WithFields(log.Fields{
						"Topic":               "Peer",
						"Key":                 fsm.pConf.State.NeighborAddress,
						"Code":                body.ErrorCode,
						"Subcode":             body.ErrorSubcode,
						"Communicated-Reason": communication,
						"Data":                rest,
					}).Warn("received notification")
					fsm.lock.RUnlock()
				} else {
					fsm.lock.RLock()
					log.WithFields(log.Fields{
						"Topic":   "Peer",
						"Key":     fsm.pConf.State.NeighborAddress,
						"Code":    body.ErrorCode,
						"Subcode": body.ErrorSubcode,
						"Data":    body.Data,
					}).Warn("received notification")
					fsm.lock.RUnlock()
				}

				fsm.lock.RLock()
				s := fsm.pConf.GracefulRestart.State
				hardReset := s.Enabled && s.NotificationEnabled && body.ErrorCode == bgp.BGP_ERROR_CEASE && body.ErrorSubcode == bgp.BGP_ERROR_SUB_HARD_RESET
				fsm.lock.RUnlock()
				if hardReset {
					sendToStateReasonCh(fsmHardReset, m)
				} else {
					sendToStateReasonCh(fsmNotificationRecv, m)
				}
				return nil, nil
			}
		}
	}
	return fmsg, nil
}

func (fsm *fsm) recvMessage() error {
	fmsg, _ := fsm.recvMessageWithError()
	if fmsg != nil {
		fsm.msgCh <- fmsg
	}
	return nil
}

func (fsm *fsm) changeadminState(s adminState) error {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()

	if fsm.adminState != s {
		log.WithFields(log.Fields{
			"Topic":      "Peer",
			"Key":        fsm.pConf.State.NeighborAddress,
			"State":      fsm.state.String(),
			"adminState": s.String(),
		}).Debug("admin state changed")

		fsm.adminState = s
		fsm.pConf.State.AdminDown = !fsm.pConf.State.AdminDown

		switch s {
		case adminStateUp:
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
			}).Info("Administrative start")
		case adminStateDown:
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
			}).Info("Administrative shutdown")
		case adminStatePfxCt:
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   fsm.pConf.State.NeighborAddress,
				"State": fsm.state.String(),
			}).Info("Administrative shutdown(Prefix limit reached)")
		}

	} else {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   fsm.pConf.State.NeighborAddress,
			"State": fsm.state.String(),
		}).Warn("cannot change to the same state")

		return fmt.Errorf("cannot change to the same state")
	}
	return nil
}
