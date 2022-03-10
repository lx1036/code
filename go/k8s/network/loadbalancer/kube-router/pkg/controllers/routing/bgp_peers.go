package routing

import (
	"context"
	"fmt"

	gobgpapi "github.com/osrg/gobgp/v3/api"
)

func (controller *NetworkRoutingController) isPeerEstablished(peerIP string) (bool, error) {
	var peerConnected bool
	err := controller.bgpServer.ListPeer(context.TODO(), &gobgpapi.ListPeerRequest{
		Address: peerIP,
	}, func(peer *gobgpapi.Peer) {
		if peer.Conf.NeighborAddress == peerIP && peer.State.SessionState == gobgpapi.PeerState_ESTABLISHED {
			peerConnected = true
		}
	})
	if err != nil {
		return false, fmt.Errorf("unable to list peers to see if tunnel & routes need to be removed: %v", err)
	}

	return peerConnected, nil
}
