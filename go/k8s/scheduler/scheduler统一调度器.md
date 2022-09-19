

# 统一调度器 coscheduler
统一调度器是既可以调度在线 Pod，也可以调度离线 Pod。
阿里的做法：就是利用 multi-profile 来做，可以设计两个 profile, online profile 和 offline profile。
offline profile 包含所有离线 Pod plugins，比如可以借鉴 volcano plugins。
文档可见：https://www.cncf.io/wp-content/uploads/2020/08/%E9%98%BF%E9%87%8C%E5%B7%B4%E5%B7%B4%E5%A6%82%E4%BD%95%E6%89%A9%E5%B1%95Kubernetes-%E8%B0%83%E5%BA%A6%E5%99%A8%E6%94%AF%E6%8C%81-AI-%E5%92%8C%E5%A4%A7%E6%95%B0%E6%8D%AE%E4%BD%9C%E4%B8%9A%EF%BC%9F1-xi-jiang.pdf

