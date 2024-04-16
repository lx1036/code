// Package workqueue features:
// - 每一个 item 被 ordered add 到队列里
// - 并发 worker 消费时，每一个 item 只会被处理一次，就算该 item 被 ordered add 多次
// - 可以有多个 consumers 和 producers, 对于已经被处理过的 item，可以被重新排队 re-enqueue
// - workqueue 可以被 shutdown notification
package workqueue
