




**[kube-controller-manager进程启动参数](https://kubernetes.io/zh/docs/reference/command-line-tools-reference/kube-controller-manager/)**
```shell script
kube-controller-manager --pod-eviction-timeout=86400s --root-ca-file=/etc/kubernetes/ssl/kube-ca.pem \
--service-cluster-ip-range=192.168.0.0/16 --allow-untagged-cloud=true \
--enable-hostpath-provisioner=false --v=2 --allocate-node-cidrs=true \
--leader-elect=true --terminated-pod-gc-threshold=1000 --cloud-provider= \
--kubeconfig=/etc/kubernetes/ssl/kubecfg-kube-controller-manager.yaml \
--service-account-private-key-file=/etc/kubernetes/ssl/kube-service-account-token-key.pem \
--profiling=false --address=0.0.0.0 --configure-cloud-routes=false \
--cluster-cidr=10.217.128.0/18 --node-monitor-grace-period=40s \
--use-service-account-credentials=true
```




# Kubernetes学习笔记之kube-controller-manager源码解析



