package reporter

import (
	"time"

	"k8s-lx1036/k8s/meetup/2020-12-31/u26/statistics"
	"k8s-lx1036/k8s/meetup/2020-12-31/u26/storage"

	log "github.com/sirupsen/logrus"
)

type Console struct {
	storage    storage.Storage
	aggregator *statistics.Aggregator
}

func (console *Console) StartRepeatedReport(stopCh <-chan struct{}, start, end time.Time) {
	for {
		select {
		case <-stopCh:
			break
		case <-time.Tick(time.Second * 10):
			// 1. 按照给定的时间区间，从数据库中拉取数据

			allRequestInfos := console.storage.GetAllRequestInfosByDuration(start, end)

			// 2. 根据原始数据，得到统计数据
			var requestsStat []statistics.RequestStat
			for _, requestInfos := range allRequestInfos {
				requestStat := console.aggregator.Aggregate(requestInfos)
				requestsStat = append(requestsStat, requestStat)
			}

			// 3. 将数据显示到终端
			for _, stat := range requestsStat {
				log.Infof("apiName: %s, count: %d, max: %v, min: %v, avg: %v", stat.ApiName,
					stat.Count, stat.MaxResponseTime, stat.MinResponseTime, stat.AvgResponseTime)
			}
		}
	}
}

func NewConsole(storage storage.Storage) Reporter {
	return &Console{
		storage:    storage,
		aggregator: statistics.NewAggregator(),
	}
}
