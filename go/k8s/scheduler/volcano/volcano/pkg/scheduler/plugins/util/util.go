package util

const (
	// Permit indicates that plugin callback function permits job to be inqueue, pipelined, or other status
	Permit = 1
	// Abstain indicates that plugin callback function abstains in voting job to be inqueue, pipelined, or other status
	Abstain = 0
	// Reject indicates that plugin callback function rejects job to be inqueue, pipelined, or other status
	Reject = -1
)
