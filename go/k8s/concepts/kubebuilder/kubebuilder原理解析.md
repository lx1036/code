

**[深入解析 Kubebuilder：让编写 CRD 变得更简单](https://zhuanlan.zhihu.com/p/83957726)**

Kubebuilder 是一个使用 CRDs 构建 K8s API 的 SDK，主要是：
提供脚手架工具初始化 CRDs 工程，自动生成 boilerplate 代码和配置；提供代码库封装底层的 K8s go-client。
方便用户从零开始开发 CRDs，Controllers 和 Admission Webhooks 来扩展 K8s。


main -> new controllerManager -> controller builder -> builder add Runnable(Controller) -> controller(Reconciler)
                                                        -> Controller start -> controller.worker
manager对象:

初始化共享cache;
初始化k8s client，用于与api-server通信
