package server

import (
	"context"
	"fmt"
	"github.com/eapache/channels"
	"github.com/osrg/gobgp/pkg/packet/bgp"
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

type fsmStateReason struct {
	Type            fsmStateReasonType
	BGPNotification *bgp.BGPMessage
	Data            []byte
}

func newFsmStateReason(typ fsmStateReasonType, notif *bgp.BGPMessage, data []byte) *fsmStateReason {
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

type fsm struct {
	gConf                *config.Global
	pConf                *config.Neighbor
	lock                 sync.RWMutex
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

type fsmHandler struct {
	fsm              *fsm
	conn             net.Conn
	msgCh            *channels.InfiniteChannel
	stateReasonCh    chan fsmStateReason
	incoming         *channels.InfiniteChannel
	outgoing         *channels.InfiniteChannel
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
	go h.loop(ctx, h.wg)
	return h
}

func (h *fsmHandler) loop(ctx context.Context, wg *sync.WaitGroup) error {

}
