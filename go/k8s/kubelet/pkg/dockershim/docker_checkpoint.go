package dockershim

type Protocol string

const (
	// default directory to store pod sandbox checkpoint files
	sandboxCheckpointDir = "sandbox"
	protocolTCP          = Protocol("tcp")
	protocolUDP          = Protocol("udp")
	protocolSCTP         = Protocol("sctp")
	schemaVersion        = "v1"
)
