
# 问题：业务pod日志通过filebeat落地kafka，配置input.yaml需要人工去添加白名单，比较麻烦
该Operator可以根据业务pod中自定义的annotation来修改input.yaml，出发filebeat重新reload，实现自动化配置。

# 参考文献
**[360 基于Kubernetes容器云日志采集与处理实践](https://xigang.github.io/2018/05/19/kubernetes-docker-log/)**
