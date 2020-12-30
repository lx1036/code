package storage

import (
	"k8s-lx1036/k8s/meetup/2020-12-31/statistics"
	"time"
)

type Storage interface {
	GetRequestInfos(start, end time.Duration)
}

type RedisMetricsStorage struct {

}

func (storage *RedisMetricsStorage) GetAllRequestInfosByDuration(start, end time.Time)  {
	mock := map[string][]statistics.RequestInfo{
		"register": {
			statistics.RequestInfo{
				ApiName:      "register",
				ResponseTime: 0,
			},
		},
		"login": {
			
		},
	}
	
	
	
}
