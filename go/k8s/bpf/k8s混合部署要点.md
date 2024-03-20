
# cpu 隔离-进程优先级调度
https://help.aliyun.com/zh/alinux/user-guide/group-identity-feature
Group Identity 功能来避免超线程干扰。


# 内存隔离-异步回收 memcg
https://help.aliyun.com/zh/alinux/user-guide/memcg-backend-asynchronous-reclaim

问题:
在社区内核系统中，系统分配内存并在相应memcg中的统计达到memcg设定的内存上限时，会触发memcg级别的直接内存回收。直接内存回收是发生在内存分配上下文的同步回收，因此会影响当前进程的性能。

方案：
为了解决这个问题，Alibaba Cloud Linux增加了memcg粒度的后台异步回收功能。
