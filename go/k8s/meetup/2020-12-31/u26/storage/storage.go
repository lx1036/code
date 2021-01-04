package storage

import (
	"time"

	"k8s-lx1036/k8s/meetup/2020-12-31/u26/request"
)

type Storage interface {
	GetRequestInfos(start, end time.Duration) []request.RequestInfo

	SaveRequestInfo(info request.RequestInfo)

	GetAllRequestInfosByDuration(start, end time.Time) map[string][]request.RequestInfo
}
