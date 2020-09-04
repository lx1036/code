
docs:
**[Workqueue机制](https://app.yinxiang.com/fx/2c3c4485-56d9-4f02-b76c-b8f04ddbe307)**


# Queue






# pkg k8s.io/client-go/util/workqueue
写一个队列Queue，实现以下功能：
(1)功能
* 有序：按照添加顺序处理元素，FIFO先进先出(first in first out)
* 去重：相同元素同一时间不会重复处理。加入到队列queue的一个job，虽加入多次但只会被process一次；同一时刻，该job不会被并发process多次。也就是说，一个job，就算加入多次，但只会被process一次。
* 并发性：有多个producers和consumers
* 标记机制：标记一个元素是否被处理过，也允许元素处理时重新排队
* 通知机制：Shutdown方法通过信号量通知队列不再接收新元素，同时通知 metric goroutine 退出
* 延迟：支持延迟一段时间再将元素入队
* 限速：可以限制元素存入队列的速率，限制一个元素被重新排队（Reenqueued）次数
* metrics: 支持 metric 监控指标，可以用于 Prometheus 监控

(2)示例
https://github.com/kubernetes/client-go/blob/master/examples/workqueue/README.md
