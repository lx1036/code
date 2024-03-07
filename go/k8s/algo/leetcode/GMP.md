

# GMP
https://iswbm.com/537.html

* G(goroutine): 一个 goroutine 最小 4k 内存，4G内存可以创建 1024*1024 goroutine，大约 1Million 个 goroutine。
goroutine 存在的意义：因为进程或线程切换时，需要从用户态到内核态的切换。  
* M(Thread): 操作系统线程，go runtime 设置最多可以创建 10K 个 thread。
* P(Processor): 机器的 cpu 核数，可是设置 GOMAXPROCS 调小。

比如，一个 4c4g pod，可以最大创建 1M 个 goroutine，10K 个 thread, 4 个 P。thread 线程数量和 cpu processor 核数数量保持一致，
这样防止线程切换非常浪费 cpu 资源，这样每一个 thread 都绑定在一个 cpu processor 上，无需线程切换。但是，也为了防止 g1 系统调用导致 thread1
处于阻塞中，p1 就需要重建一个 thread2 来继续运行其他 goroutine。所以，不是那么死板。

所以，现在有 1M 个 goroutine 等着被调度到 4 个 thread/processor 上。

## Goroutine 调度

goroutine 所在的队列：
* local queue: 每一个 thread 都有一个自己的 local queue，当某个 goroutine 找到可以绑定的 thread，就会存放在该 thread 所属的 local queue。
* global queue: 全局 queue，当某个 goroutine 没有找到对应的 cpu 可以调度(说明 4 个 cpu 都在处理各自 local queue 的 goroutine，很忙)，
那这个 goroutine 就会放到 global queue。
