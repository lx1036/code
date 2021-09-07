package raft

type Node struct {
	ID   int    `json:"id"`
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

type Config struct {
	Node  *Node   `json:"node"`  // local node
	Nodes []*Node `json:"nodes"` // nodes cluster, include local node

	HeartBeatTime int `json:"heartbeattime"`
}

func (config *Config) GetPeers() Peers {
	var peers Peers
	for _, node := range config.Nodes {
		if node.ID != config.Node.ID {
			peers = append(peers, &Peer{
				Node: node,
			})
		}
	}

	return peers
}
