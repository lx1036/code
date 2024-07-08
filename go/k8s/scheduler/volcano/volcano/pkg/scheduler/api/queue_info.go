package api

import "k8s.io/apimachinery/pkg/types"

type QueueID types.UID

type QueueInfo struct {
	UID QueueID
}
