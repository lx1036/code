package reporter

import (
	"time"

	"k8s-lx1036/k8s/meetup/2020-12-31/u26/statistics"
	"k8s-lx1036/k8s/meetup/2020-12-31/u26/storage"
)

type Email struct {
	ToAddress  []string
	storage    storage.Storage
	aggregator *statistics.Aggregator
}

func (email *Email) AddToAddress(address []string) {
	email.ToAddress = append(email.ToAddress, address...)
}

func (email *Email) StartRepeatedReport(stopCh <-chan struct{}, start, end time.Time) {
	for {
		select {
		case <-stopCh:
			break

		case <-time.Tick(time.Second * 10):
			// email 业务逻辑

			// 1. 按照给定的时间区间，从数据库中拉取数据

			allRequestInfos := email.storage.GetAllRequestInfosByDuration(start, end)

			// 2. 根据原始数据，得到统计数据
			var requestsStat []statistics.RequestStat
			for _, requestInfos := range allRequestInfos {
				requestStat := email.aggregator.Aggregate(requestInfos)
				requestsStat = append(requestsStat, requestStat)
			}

			// 3. 将数据通过邮件发出
			// ...
		}
	}
}

func NewMail(ToAddress []string, storage storage.Storage, aggregator *statistics.Aggregator) Reporter {
	return &Email{
		ToAddress:  ToAddress,
		storage:    storage,
		aggregator: aggregator,
	}
}
