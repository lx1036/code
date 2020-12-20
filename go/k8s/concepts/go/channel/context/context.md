
**[Go Concurrency Patterns: Context](https://blog.golang.org/context)**
**[Concurrency is not Parallelism](https://talks.golang.org/2012/waza.slide#1)**
**[6.1 上下文 Context](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-context/)**
**[proposal: context: new package for standard library](https://github.com/golang/go/issues/14660)**

**[golang context的一些思考](https://tech.ipalfish.com/blog/2020/03/30/golang-context/)**

# Context 设计目的
Golang核心库context的设计目的和使用，context库的设计目的主要是跟踪goroutine调用树，并在树中传递通知和数据:
* (1)退出/过期通知，可以链式给树中每一个goroutine传递退出机制，集体退出。
* (2)传递数据，可以给树中每一个goroutine传递数据。

学习golang关于context库官方博客：**[Go Concurrency Patterns: Context](https://blog.golang.org/context)**，并结合Gin的context struct学习context的使用(https://github.com/gin-gonic/gin/blob/master/context.go)。
cacelctx:

timerctx:

valuectx:
