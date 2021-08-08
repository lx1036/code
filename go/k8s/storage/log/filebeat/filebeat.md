
# Filebeat

## Modules
modules 简化了日志收集、解析，有 fileset 组成，比如 nginx fileset 包含 access/error/ingress-controller 等 fileset。
主要就是：把常见的一些软件日志收集，做成一个模块给用户直接配置使用。比如redis module，可以收集redis logs，和 redis slow keys 等这些功能，
用户只需要直接配置使用就好，不需要考虑其他东西。或者像 nginx module，nginx log 有一些特定格式，也只需配置使用 nginx module 就行。这大大
减少用户使用成本。

