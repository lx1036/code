package storage

import (
	"time"

	"k8s-lx1036/k8s/meetup/2020-12-31/u26/request"
)

type RedisMetricsStorage struct {
}

func (storage *RedisMetricsStorage) SaveRequestInfo(info request.RequestInfo) {

}

func (storage *RedisMetricsStorage) GetRequestInfos(start, end time.Duration) []request.RequestInfo {
	return nil
}

func (storage *RedisMetricsStorage) GetAllRequestInfosByDuration(start, end time.Time) map[string][]request.RequestInfo {
	mock := map[string][]request.RequestInfo{
		"register": {
			request.RequestInfo{
				ApiName:      "register",
				ResponseTime: time.Second * 1,
				Timestamp:    time.Now(),
			},
			request.RequestInfo{
				ApiName:      "register",
				ResponseTime: time.Second * 2,
				Timestamp:    time.Now().Add(time.Second * 30),
			},
			request.RequestInfo{
				ApiName:      "register",
				ResponseTime: time.Second * 3,
				Timestamp:    time.Now().Add(time.Second * 60),
			},
		},
		"login": {
			request.RequestInfo{
				ApiName:      "login",
				ResponseTime: time.Second * 1,
				Timestamp:    time.Now().Add(time.Second),
			},
			request.RequestInfo{
				ApiName:      "login",
				ResponseTime: time.Second * 2,
				Timestamp:    time.Now().Add(time.Second * 30),
			},
		},
	}

	return mock
}

func NewRedisMetricsStorage() *RedisMetricsStorage {
	return &RedisMetricsStorage{}
}
