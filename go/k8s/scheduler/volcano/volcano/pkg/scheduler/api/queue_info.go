package api

import (
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling"
)

type QueueID types.UID

type QueueInfo struct {
	UID  QueueID
	Name string

	Queue *scheduling.Queue
}
