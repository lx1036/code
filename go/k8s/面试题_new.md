

# 基础
(1)select 和 epoll区别？


(2)HTTP 和 HTTPS 的区别?TCP 拥塞控制?HTTP 长连接，短连接? 
tcp、udp区别? 描述一下 TCP 四次挥手的过程？TCP 有哪些状态？TCP 的 LISTEN 状态是什么？
TCP 的 CLOSE_WAIT 状态是什么？建立一个 socket 连接要经过哪些步骤？
常见的 HTTP 状态码有哪些？301和302有什么区别？504和500有什么区别？
TCP和UDP的区别，Tcp的拥塞控制，Tcp的流量控制，拥塞控制算法，Tcp三次握手？
快速重传？快恢复会进入哪个阶段？
详细描述HTTPS的加密过程，需要几次通信？
三次握手和四次挥手，说一下time_wait?


(3)用过哪些锁，自旋锁和互斥锁有什么区别？用过哪些分布式锁。答了 mysql，redis， zookeeper 分别聊了一下优缺点。
redis setnx + expire 有什么缺点，如何优化？


(4)打开一个 URL 的过程。（这个也是必考点，基本上每个人都会，所以尽量说点其他人不知道的，一定要有自己的思考，让面试官眼前一亮，而不是觉得你在背书）


# 算法










# golang
建议：
* 比如 Go 的 GMP 模型，垃圾回收，channel 都是必考点，最好去读一下源码。



(1)(头条面试题)GMP 调度模型，很多面试官都会问这个，一定要好好复习，要讲出亮点，讲出其他同学讲不出的东西？
典藏版Golang调度器GPM原理与调度全分析: https://www.jianshu.com/p/fa696563c38a



(2)(头条面试题)Context 的用法?


(3)Go的内存模型是什么？如何解决Go的内存泄漏的问题？


(4)(头条面试题)go的垃圾回收以及调度模型?go的垃圾回收，哪种机制?好在哪里，不好在哪里？
典藏版Golang三色标记、混合写屏障GC模式图文全分析: https://www.jianshu.com/p/4c5a303af470


(5)描述一下go的协程实现?


(6)goroutine 是怎么调度的？goroutine 和 kernel thread 之间是什么关系？


(7)(头条面试题)Go 的逃逸分析了解过吗，能不能写一个?




# 中间件
建议：
* 中间件：这个也是必问的一个环节，尽量要多了解一些，但是一定要说自己会的，至少知道运行原理和特点。比较重要的中间件有：kafka。



(1)(头条面试题)Kafka 的消费者如何做消息去重?介绍一下 Kafka 的 ConsumerGroup？



# 数据库
建议：
* 数据库和网络一定要重点复习。最重要的肯定是 Mysql，Mysql中比较重要的就是隔离级别和索引，一定一定要弄懂。然后就是 redis，也是经常会问的一个东西。
* redis所有面试题，包含答案：https://leetcode-cn.com/circle/article/pLsmO2/



(1)(头条面试题)mysql幻读是怎么情况，如何避免的？


(2)(头条面试题)B树和B+树的区别，为什么mysql要用B+树，mongodb要用B树。


(3)(头条面试题)redis的跳表知道吗，为什么不用红黑树。我回答了因为红黑树实现比跳表复杂。


(4)(头条面试题)Mysql 集群如何保证数据的一致性。分别回答了弱一致性和强一致性


(5)[延迟队列]使用过 Redis 做异步队列么，你是怎么用的？redis如何实现延时队列？(学习时，记得参考k8s client-go workqueue包的delaying_queue)
延时队列：延迟队列用途，比如我指定本技术文章在下周一发布，就需要把这篇文章的id和time消息加入延迟队列中，下周一才会pop item，进入文章发布程序。
使用sortedset，拿时间戳作为 score，消息内容ID作为 key 调用 zadd 来生产消息，消费者用 zrangebyscore 指令获取 N 秒之前的数据轮询进行处理。
```shell
# https://medium.com/@cheukfung/redis%E5%BB%B6%E8%BF%9F%E9%98%9F%E5%88%97-c940850a264f
# https://redis.io/commands/zadd zadd key score member [score member...]
# 消费者通过ZRANGEBYSCORE获取消息。如果时间未到，将得不到消息；当时间已到或已超时，都可以得到消息
ZADD delay-queue 1520985600 "publish article"
ZRANGEBYSCORE delay-queue -inf 1520985599 WITHSCORES
# (empty array)
ZRANGEBYSCORE delay-queue -inf 1520985600 WITHSCORES
# 1) "publish article"
# 2) "1520985600"

# 使用ZRANGEBYSCORE取得消息后，消息并没有从集合中删出，需要调用ZREM删除消息
ZREM delay-queue "publish article"
```

golang实现delaying_queue: https://mp.weixin.qq.com/s/aZC2MXQFuUu00TgmOozwKQ




(6)[分布式锁]使用过 Redis 分布式锁么，它是什么回事？如果在 setnx 之后执行 expire 之前进程意外 crash 或者要重启维护了，那会怎么样？
先拿 setnx 来争抢锁，抢到之后，再用 expire 给锁加一个过期时间防止锁忘记了释放。
set 指令有非常复杂的参数，这个应该是可以同时把 setnx 和 expire 合成一条指令来用的！





# 容器云
(1) k8s 的 watch 机制如何保证数据不丢失？




