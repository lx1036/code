

# Evict BE Pod

## 背景
尽管有 cpu.cfs_quota_us 设置整机水位线对 BE pod cpu 压制，但是频繁压制不利于 BE pod 工作，所以需要设置一个驱逐水位线 evict-threshold，
达到了驱逐水位线就驱逐一些 BE pod，给其他 BE pod 腾出一定的 cpu 资源。这个功能不需要设置 cpu cgroup。


## 验证







## 参考文献
