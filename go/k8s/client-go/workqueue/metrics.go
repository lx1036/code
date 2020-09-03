package workqueue

type queueMetrics interface {
	add(item job)
	get(item job)
	done(item job)
	updateUnfinishedWork()
}
