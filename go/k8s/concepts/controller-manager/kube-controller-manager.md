


**[Kube-Controller-manager之StatefulSet Controller源码解析](https://xigang.github.io/2019/12/27/statefulset-controller/)**
**[Kube-Controller-manager之Replicaset Controller源码解析](https://xigang.github.io/2018/09/16/replicaset/)**
**[Kube-Controller-manager之Deployment Controller源码解析](https://xigang.github.io/2018/09/08/deployment/)**


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
