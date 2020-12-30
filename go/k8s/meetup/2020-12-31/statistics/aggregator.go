package statistics

import "time"

type RequestStat struct {
	maxResponseTime
	minResponseTime
	
	p99ResponseTime
	p999ResponseTime
	
	avgResponseTime
	
	count
	tps
}


// 根据原始数据，得到统计数据
type Aggregator struct {

}

func (aggregator *Aggregator) Aggregate(requestInfo RequestInfo) RequestStat {

}
